package telegram

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.telegram.org",
	}
}

// Ping checks Telegram API connectivity using getMe
func (c *Client) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/bot%s/getMe", c.baseURL, c.token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}

	return nil
}
