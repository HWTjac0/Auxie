package handlers

import (
	"auxie/backend/internal/models"
	"auxie/backend/internal/repositories"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SpotifyHandler struct {
	userRepo repositories.UserRepository
}

func NewSpotifyHandler(userRepo repositories.UserRepository) *SpotifyHandler {
	return &SpotifyHandler{userRepo: userRepo}
}

// MOST OF THIS CODE IS TEMPORARY FOR TESTING PURPOSES
type SpotifyTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type SpotifyUserResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Images      []struct {
		URL string `json:"url"`
	} `json:"images"`
}

func (h *SpotifyHandler) SpotifyLogin(c *gin.Context) {
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

func (h *SpotifyHandler) SpotifyCallback(c *gin.Context) {
	code := c.Query("code")
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No authorization code"})
		return
	}

	redirectURI := "http://127.0.0.1:8080/api/v1/auth/spotify/callback"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	authHeader := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error exchanging code for token"})
		return
	}
	defer resp.Body.Close()

	var tokenData SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding token"})
		return
	}

	reqMe, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)

	respMe, err := client.Do(reqMe)
	if err != nil || respMe.StatusCode != 200 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error loading profile"})
		return
	}
	defer respMe.Body.Close()

	var userResponse SpotifyUserResponse
	if err := json.NewDecoder(respMe.Body).Decode(&userResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding user data"})
		return
	}

	var dbUser *models.User

	existingUser, err := h.userRepo.GetBySpotifyID(userResponse.ID)
	if err == nil && existingUser != nil {
		dbUser = existingUser
	} else {
		session := sessions.Default(c)
		sessionUserID := session.Get("user_id")

		if sessionUserID != nil {
			var guestUserID int
			switch val := sessionUserID.(type) {
			case int:
				guestUserID = val
			case int64:
				guestUserID = int(val)
			case float64:
				guestUserID = int(val)
			}

			if guestUserID > 0 {
				dbUser = &models.User{ID: guestUserID}
			}
		}

		if dbUser == nil {
			newUser := &models.User{
				Username:  userResponse.DisplayName,
				Type:      models.UserTypeRegistered,
				CreatedAt: time.Now(),
			}
			newID, err := h.userRepo.Create(newUser)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
			dbUser = &models.User{ID: int(newID)}
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
	if len(userResponse.Images) > 0 {
		session.Set("user_image", userResponse.Images[0].URL)
	} else {
		session.Set("user_image", "")
	}

	if err := session.Save(); err != nil {
		fmt.Println("Error saving session:", err)
	}

	c.Redirect(http.StatusTemporaryRedirect, "http://127.0.0.1:5173/")
}
