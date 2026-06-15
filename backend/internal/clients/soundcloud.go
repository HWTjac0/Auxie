package clients

import (
	"net/url"
	"strings"
)

type SoundCloudClient struct {
	base         *BaseClient
	clientID     string
	clientSecret string
	redirectURI  string
}

type SoundCloudTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func NewSoundCloudClient(clientID, clientSecret, redirectURI string) *SoundCloudClient {
	return &SoundCloudClient{
		base:         NewBaseClient(),
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

func (c *SoundCloudClient) TokenRefresh(refreshToken string) (*SoundCloudTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("refresh_token", refreshToken)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Accept":       "application/json",
	}

	var tokenResp SoundCloudTokenResponse
	err := c.base.Request("POST", "https://api.soundcloud.com/oauth2/token", headers, strings.NewReader(data.Encode()), &tokenResp)
	if err != nil {
		return nil, err
	}
	return &tokenResp, nil
}
