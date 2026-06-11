package handlers

import (
	"auxie/backend/internal/models"
	"auxie/backend/internal/repositories"
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

type SoundCloudHandler struct {
	userRepo repositories.UserRepository
}

func NewSoundCloudHandler(userRepo repositories.UserRepository) *SoundCloudHandler {
	return &SoundCloudHandler{userRepo: userRepo}
}

type SoundCloudTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type SoundCloudUserResponse struct {
	ID        interface{} `json:"id"`
	Username  string      `json:"username"`
	AvatarURL string      `json:"avatar_url"`
}

func (h *SoundCloudHandler) SoundCloudLogin(c *gin.Context) {
	clientID := os.Getenv("SOUNDCLOUD_CLIENT_ID")
	redirectURI := "http://127.0.0.1:8080/api/v1/auth/soundcloud/callback"

	authURL := "https://api.soundcloud.com/connect?" + url.Values{
		"response_type": {"code"},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
	}.Encode()

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *SoundCloudHandler) SoundCloudCallback(c *gin.Context) {
	code := c.Query("code")
	clientID := os.Getenv("SOUNDCLOUD_CLIENT_ID")
	clientSecret := os.Getenv("SOUNDCLOUD_CLIENT_SECRET")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No authorization code"})
		return
	}

	redirectURI := "http://127.0.0.1:8080/api/v1/auth/soundcloud/callback"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, _ := http.NewRequest("POST", "https://api.soundcloud.com/oauth2/token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error exchanging code for token"})
		return
	}
	defer resp.Body.Close()

	var tokenData SoundCloudTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding token"})
		return
	}

	reqMe, _ := http.NewRequest("GET", "https://api.soundcloud.com/me", nil)
	reqMe.Header.Set("Authorization", "OAuth "+tokenData.AccessToken)

	respMe, err := client.Do(reqMe)
	if err != nil || respMe.StatusCode != 200 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error loading profile"})
		return
	}
	defer respMe.Body.Close()

	var userResponse SoundCloudUserResponse
	if err := json.NewDecoder(respMe.Body).Decode(&userResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding user data"})
		return
	}

	soundCloudID := fmt.Sprintf("%v", userResponse.ID)
	var dbUser *models.User

	existingUser, err := h.userRepo.GetBySoundCloudID(soundCloudID)
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
				Username:  userResponse.Username,
				Type:      models.UserTypeRegistered,
				CreatedAt: time.Now(),
			}
			if newUser.Username == "" {
				newUser.Username = "SoundCloudUser"
			}
			newID, err := h.userRepo.Create(newUser)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
			dbUser = &models.User{ID: int(newID)}
		}
	}

	err = h.userRepo.UpdateSoundCloudInfo(dbUser.ID, soundCloudID, tokenData.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SoundCloud info in database"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", dbUser.ID)
	if userResponse.Username != "" {
		session.Set("user_name", userResponse.Username)
		session.Set("soundcloud_name", userResponse.Username)
	}
	currentImage := session.Get("user_image")
	if (currentImage == nil || currentImage == "") && userResponse.AvatarURL != "" {
		session.Set("user_image", userResponse.AvatarURL)
	}

	if err := session.Save(); err != nil {
		fmt.Println("Error saving session:", err)
	}

	c.Redirect(http.StatusTemporaryRedirect, "http://127.0.0.1:5173/")
}
