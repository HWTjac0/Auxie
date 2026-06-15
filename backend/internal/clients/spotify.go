package clients

import (
	"encoding/base64"
	"net/url"
	"strings"
)

type SpotifyClient struct {
	base         *BaseClient
	clientID     string
	clientSecret string
	redirectURI  string
}

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

func NewSpotifyClient(clientID, clientSecret, redirectURI string) *SpotifyClient {
	return &SpotifyClient{
		base:         NewBaseClient(),
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

// ExchangeCode wymienia kod OAuth2 na tokeny.
func (c *SpotifyClient) ExchangeCode(code string) (*SpotifyTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.redirectURI)

	authHeader := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + authHeader,
	}

	var tokenResp SpotifyTokenResponse
	err := c.base.Request("POST", "https://accounts.spotify.com/api/token", headers, strings.NewReader(data.Encode()), &tokenResp)
	if err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func (c *SpotifyClient) TokenRefresh(refreshToken string) (*SpotifyTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	authHeader := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + authHeader,
	}

	var tokenResp SpotifyTokenResponse
	err := c.base.Request("POST", "https://accounts.spotify.com/api/token", headers, strings.NewReader(data.Encode()), &tokenResp)
	if err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func (c *SpotifyClient) GetUserProfile(accessToken string) (*SpotifyUserResponse, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	var userResp SpotifyUserResponse
	err := c.base.Request("GET", "https://api.spotify.com/v1/me", headers, nil, &userResp)
	if err != nil {
		return nil, err
	}
	return &userResp, nil
}

// SearchTrack wyszukuje utwory w Spotify.
func (c *SpotifyClient) SearchTrack(accessToken, query string) (interface{}, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	rawURL := "https://api.spotify.com/v1/search?type=track&market=PL&q=" + url.QueryEscape(query)

	var result interface{}
	err := c.base.Request("GET", rawURL, headers, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
