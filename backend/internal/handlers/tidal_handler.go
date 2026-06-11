package handlers

import (
	"auxie/backend/internal/models"
	"auxie/backend/internal/repositories"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}

func generatePKCE() (string, string, error) {
	verifier, err := generateRandomString(64)
	if err != nil {
		return "", "", err
	}

	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge, nil
}

type TidalHandler struct {
	userRepo repositories.UserRepository
}

func NewTidalHandler(userRepo repositories.UserRepository) *TidalHandler {
	return &TidalHandler{userRepo: userRepo}
}

type TidalTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type TidalUserResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Username  string `json:"username"`
			Email     string `json:"email"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"attributes"`
	} `json:"data"`
}

func (h *TidalHandler) Login(c *gin.Context) {
	clientID := os.Getenv("TIDAL_CLIENT_ID")
	redirectURI := "http://127.0.0.1:8080/api/v1/auth/tidal/callback"

	verifier, challenge, err := generatePKCE()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PKCE"})
		return
	}

	state, err := generateRandomString(32)
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
		"scope":                 {"user.read"},
		"code_challenge_method": {"S256"},
		"code_challenge":        {challenge},
		"state":                 {state},
	}.Encode()

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *TidalHandler) Callback(c *gin.Context) {
	// It need tidying
	code := c.Query("code")
	state := c.Query("state")
	clientID := os.Getenv("TIDAL_CLIENT_ID")

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

	redirectURI := "http://127.0.0.1:8080/api/v1/auth/tidal/callback"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier.(string))

	req, _ := http.NewRequest("POST", "https://auth.tidal.com/v1/oauth2/token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error exchanging code for token"})
		return
	}
	defer resp.Body.Close()

	var tokenData TidalTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding token"})
		return
	}

	reqMe, _ := http.NewRequest("GET", "https://openapi.tidal.com/v2/users/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
	reqMe.Header.Set("Accept", "application/vnd.api+json")

	respMe, err := client.Do(reqMe)
	if err != nil || respMe.StatusCode != 200 {
		bodybytes, err := io.ReadAll(respMe.Body)
		var body string
		if err != nil {
			body = "Error loading"
		} else {
			body = string(bodybytes)
		}
		c.JSON(respMe.StatusCode, gin.H{"error": body})
		return
	}
	defer respMe.Body.Close()

	var userResponse TidalUserResponse
	if err := json.NewDecoder(respMe.Body).Decode(&userResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding user data"})
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
				Username:  displayName,
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

	err = h.userRepo.UpdateTidalInfo(dbUser.ID, tidalID, tokenData.AccessToken)
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
