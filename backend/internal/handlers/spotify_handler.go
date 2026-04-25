package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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

func SpotifyLogin(c *gin.Context) {
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

func SpotifyCallback(c *gin.Context) {
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

	session := sessions.Default(c)
	fmt.Println(userResponse.DisplayName)
	session.Set("user_name", userResponse.DisplayName)
	session.Set("user_id", userResponse.ID)

	if len(userResponse.Images) > 0 {
		session.Set("user_image", userResponse.Images[0].URL)
	}
	if err := session.Save(); err != nil {
		fmt.Println("Error saving session")
	}

	c.Redirect(http.StatusTemporaryRedirect, "http://127.0.0.1:5173/dashboard")
}

func GetMe(c *gin.Context) {
	session := sessions.Default(c)
	name := session.Get("user_name")

	if name == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No active session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":  name,
		"id":    session.Get("user_id"),
		"image": session.Get("user_image"),
	})
}
