package clients

import (
	"fmt"
	"net/url"
	"strings"
)

type TidalClient struct {
	base        *BaseClient
	clientID    string
	redirectURI string
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

func NewTidalClient(clientID, redirectURI string) *TidalClient {
	return &TidalClient{
		base:        NewBaseClient(),
		clientID:    clientID,
		redirectURI: redirectURI,
	}
}

// ExchangeCode wymienia kod OAuth2 na tokeny (wymaga code_verifier dla PKCE).
func (c *TidalClient) ExchangeCode(code, codeVerifier string) (*TidalTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.clientID)
	data.Set("code", code)
	data.Set("redirect_uri", c.redirectURI)
	data.Set("code_verifier", codeVerifier)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	var tokenResp TidalTokenResponse
	err := c.base.Request("POST", "https://auth.tidal.com/v1/oauth2/token", headers, strings.NewReader(data.Encode()), &tokenResp)
	if err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func (c *TidalClient) TokenRefresh(refreshToken string) (*TidalTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", c.clientID)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	var tokenResp TidalTokenResponse
	err := c.base.Request("POST", "https://auth.tidal.com/v1/oauth2/token", headers, strings.NewReader(data.Encode()), &tokenResp)
	if err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

// GetUserProfile pobiera profil użytkownika.
func (c *TidalClient) GetUserProfile(accessToken string) (*TidalUserResponse, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/vnd.api+json",
	}

	var userResp TidalUserResponse
	err := c.base.Request("GET", "https://openapi.tidal.com/v2/users/me", headers, nil, &userResp)
	if err != nil {
		return nil, err
	}
	return &userResp, nil
}

func (c *TidalClient) SearchTrack(accessToken, query string) (interface{}, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/vnd.api+json",
	}

	parsedUrl, err := url.Parse("https://openapi.tidal.com")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}
	parsedUrl = parsedUrl.JoinPath("v2", "searchResults", query)

	q := parsedUrl.Query()
	q.Set("countryCode", "PL")
	q.Set("explicitFilter", "INCLUDE")
	q.Set("include", "tracks,albums,tracks.artists,tracks.albums")
	q.Set("page[limit]", "5")

	parsedUrl.RawQuery = q.Encode()

	var result interface{}
	err = c.base.Request("GET", parsedUrl.String(), headers, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
