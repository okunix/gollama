package gollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	token  string
	host   string
	client *http.Client
}

type clientOption func(c *Client) error

func NewClient(ctx context.Context, opts ...clientOption) (*Client, error) {
	client := Client{client: http.DefaultClient}
	for _, opt := range opts {
		if err := opt(&client); err != nil {
			return nil, err
		}
	}
	err := client.Ping(ctx)
	return &client, err
}

func WithToken(token string) clientOption {
	return func(c *Client) error {
		c.token = token
		return nil
	}
}

func WithHost(host string) clientOption {
	return func(c *Client) error {
		c.host = host
		return nil
	}
}

func WithTimeout(timeout time.Duration) clientOption {
	return func(c *Client) error {
		c.client.Timeout = timeout
		return nil
	}
}

func (c *Client) Version(ctx context.Context) (string, error) {
	type response struct {
		Version string `json:"version"`
	}

	url := c.host + "/api/version"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get ollama version: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get ollama version: %w", err)
	}

	var getVersionResponse response

	err = json.NewDecoder(resp.Body).Decode(&getVersionResponse)
	if err != nil {
		return "", fmt.Errorf("failed to decode response object: %w", err)
	}

	return getVersionResponse.Version, nil
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.host, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to ping server: not ok status code")
	}
	return nil
}
