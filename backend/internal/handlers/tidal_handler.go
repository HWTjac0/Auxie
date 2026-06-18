package handlers

import (
	"auxie/backend/internal/clients"
	"auxie/backend/internal/models"
	"auxie/backend/internal/repositories"
	"auxie/backend/internal/utils"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type TidalHandler struct {
	userRepo    repositories.UserRepository
	tidalClient *clients.TidalClient
}

func NewTidalHandler(userRepo repositories.UserRepository, tidalClient *clients.TidalClient) *TidalHandler {
	return &TidalHandler{userRepo: userRepo, tidalClient: tidalClient}
}

func (h *TidalHandler) Login(c *gin.Context) {
	clientID := os.Getenv("TIDAL_CLIENT_ID")
	redirectURI := "http://127.0.0.1:8080/api/v1/auth/tidal/callback"

	verifier, challenge, err := utils.GeneratePKCE()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PKCE"})
		return
	}

	state, err := utils.GenerateRandomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	session := sessions.Default(c)
	session.Set("tidal_pkce_verifier", verifier)
	session.Set("tidal_oauth_state", state)
	session.Save()

	authURL := "https://login.tidal.com/authorize?" + url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {"user.read search.read playback"},
		"code_challenge_method": {"S256"},
		"code_challenge":        {challenge},
		"state":                 {state},
	}.Encode()

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *TidalHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No authorization code"})
		return
	}

	session := sessions.Default(c)
	savedState := session.Get("tidal_oauth_state")
	if savedState == nil || state != savedState.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	codeVerifier := session.Get("tidal_pkce_verifier")
	if codeVerifier == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code verifier not found"})
		return
	}

	tokenData, err := h.tidalClient.ExchangeCode(code, codeVerifier.(string))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error exchanging code for token: " + err.Error()})
		return
	}

	userResponse, err := h.tidalClient.GetUserProfile(tokenData.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error loading profile: " + err.Error()})
		return
	}

	tidalID := userResponse.Data.ID
	if tidalID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tidal user ID is empty"})
		return
	}

	displayName := strings.TrimSpace(userResponse.Data.Attributes.FirstName + " " + userResponse.Data.Attributes.LastName)
	if displayName == "" {
		displayName = userResponse.Data.Attributes.Username
	}
	if displayName == "" {
		displayName = "TidalUser"
	}

	var dbUser *models.User

	existingUser, err := h.userRepo.GetByTidalID(tidalID)
	if err == nil && existingUser != nil {
		dbUser = existingUser
	} else {
		dbUser, err = getOrCreateUser(c, h.userRepo, displayName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve or create user"})
			return
		}
	}

	expiresAt := time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second)
	err = h.userRepo.UpdateTidalInfo(dbUser.ID, tidalID, tokenData.AccessToken, tokenData.RefreshToken, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Tidal info in database"})
		return
	}

	session = sessions.Default(c)
	session.Set("user_id", dbUser.ID)
	session.Set("user_name", displayName)
	session.Set("tidal_name", displayName)

	if err := session.Save(); err != nil {
		fmt.Println("Error saving session:", err)
	}

	c.Redirect(http.StatusTemporaryRedirect, "http://127.0.0.1:5173/")
}
