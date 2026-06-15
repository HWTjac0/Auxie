package handlers

import (
	"auxie/backend/internal/clients"
	"auxie/backend/internal/models"
	"auxie/backend/internal/repositories"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SpotifyHandler struct {
	userRepo      repositories.UserRepository
	spotifyClient *clients.SpotifyClient
}

func NewSpotifyHandler(userRepo repositories.UserRepository, spotifyClient *clients.SpotifyClient) *SpotifyHandler {
	return &SpotifyHandler{userRepo: userRepo, spotifyClient: spotifyClient}
}

func (h *SpotifyHandler) Login(c *gin.Context) {
	scope := "user-read-private user-read-email user-modify-playback-state"
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	redirectURI := "http://127.0.0.1:8080/api/v1/auth/spotify/callback"

	authURL := "https://accounts.spotify.com/authorize?" + url.Values{
		"response_type": {"code"},
		"client_id":     {clientID},
		"scope":         {scope},
		"redirect_uri":  {redirectURI},
		"show_dialog":   {"true"},
	}.Encode()

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *SpotifyHandler) Callback(c *gin.Context) {
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No authorization code"})
		return
	}

	tokenData, err := h.spotifyClient.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error exchanging code for token"})
		return
	}

	userResponse, err := h.spotifyClient.GetUserProfile(tokenData.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error loading profile"})
		return
	}

	var dbUser *models.User

	existingUser, err := h.userRepo.GetBySpotifyID(userResponse.ID)
	if err == nil && existingUser != nil {
		dbUser = existingUser
	} else {
		dbUser, err = getOrCreateUser(c, h.userRepo, userResponse.DisplayName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve or create user"})
			return
		}
	}

	expiresAt := time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second)
	err = h.userRepo.UpdateSpotifyInfo(dbUser.ID, userResponse.ID, tokenData.AccessToken, tokenData.RefreshToken, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Spotify info in database"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", dbUser.ID)
	session.Set("user_name", userResponse.DisplayName)
	session.Set("spotify_name", userResponse.DisplayName)
	currentImage := session.Get("user_image")
	if (currentImage == nil || currentImage == "") && len(userResponse.Images) > 0 {
		session.Set("user_image", userResponse.Images[0].URL)
	}

	if err := session.Save(); err != nil {
		fmt.Println("Error saving session:", err)
	}

	c.Redirect(http.StatusTemporaryRedirect, "http://127.0.0.1:5173/")
}

func (h *SpotifyHandler) GetUserAccessToken(c *gin.Context) (string, error) {
	userId := c.GetInt("user_id")
	dbUser, err := h.userRepo.GetByID(userId)
	if err != nil {
		return "", err
	}

	return dbUser.SpotifyAuthKey.String, nil
}

func (h *SpotifyHandler) SearchTrack(c *gin.Context) {
	accessToken, err := h.GetUserAccessToken(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	keywords := c.Query("search")
	result, err := h.spotifyClient.SearchTrack(accessToken, keywords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resp": result})
}
