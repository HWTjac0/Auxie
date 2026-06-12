package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BaseClient struct {
	client *http.Client
}

func NewBaseClient() *BaseClient {
	return &BaseClient{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Request performs an HTTP request, checks the status, and automatically decodes the response into the target struct.
func (c *BaseClient) Request(method, rawURL string, headers map[string]string, body io.Reader, target interface{}) error {
	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
