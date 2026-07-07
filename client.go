package gollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func (c *Client) parseError(resp *http.Response) error {
	var ollamaErr Error
	err := json.NewDecoder(resp.Body).Decode(&ollamaErr)
	if err != nil {
		return fmt.Errorf("ollama error status code: %d", resp.StatusCode)
	}
	return fmt.Errorf("ollama error response: %w", error(ollamaErr))
}

func (c *Client) newRequest(
	ctx context.Context,
	method, url string,
	body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if len(c.token) > 0 {
		req.Header.Add("Authorization", "Bearer "+c.token)
	}
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := c.newRequest(ctx, "GET", c.host, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}
	return nil
}

func (c *Client) Version(ctx context.Context) (string, error) {
	type response struct {
		Version string `json:"version"`
	}

	url := c.host + "/api/version"

	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get ollama version: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", c.parseError(resp)
	}

	var getVersionResponse response

	err = json.NewDecoder(resp.Body).Decode(&getVersionResponse)
	if err != nil {
		return "", err
	}

	return getVersionResponse.Version, nil
}

func (c *Client) Tags(ctx context.Context) ([]Model, error) {
	type response struct {
		Models []Model `json:"models"`
	}
	url := c.host + "/api/tags"
	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var tagsResponse response
	err = json.NewDecoder(resp.Body).Decode(&tagsResponse)
	if err != nil {
		return nil, err
	}

	return tagsResponse.Models, nil
}

func (c *Client) Ps(ctx context.Context) ([]Ps, error) {
	type response struct {
		Models []Ps `json:"models"`
	}
	url := c.host + "/api/ps"
	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var psResponse response
	err = json.NewDecoder(resp.Body).Decode(&psResponse)
	if err != nil {
		return nil, err
	}

	return psResponse.Models, nil
}

func (c *Client) toBody(a any) (io.Reader, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func (c *Client) ShowModelDetails(
	ctx context.Context,
	model string,
	verbose bool,
) (ModelDetails, error) {
	type request struct {
		Model   string `json:"model"`
		Verbose bool   `json:"verbose"`
	}
	url := c.host + "/api/show"
	body, _ := c.toBody(request{Model: model, Verbose: verbose})
	req, err := c.newRequest(ctx, "POST", url, body)
	if err != nil {
		return ModelDetails{}, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return ModelDetails{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return ModelDetails{}, c.parseError(resp)
	}

	var detailsResponse ModelDetails
	err = json.NewDecoder(resp.Body).Decode(&detailsResponse)
	if err != nil {
		return ModelDetails{}, err
	}

	return detailsResponse, nil
}

func (c *Client) Delete(ctx context.Context, model string) error {
	type request struct {
		Model string `json:"model"`
	}
	url := c.host + "/api/delete"
	body, _ := c.toBody(request{Model: model})
	req, err := c.newRequest(ctx, "DELETE", url, body)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}
	return nil
}

func (c *Client) Create(ctx context.Context, model CreateModel) error {
	model.Stream = false
	url := c.host + "/api/create"
	body, _ := c.toBody(model)
	req, err := c.newRequest(ctx, "POST", url, body)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}
	return nil
}

func (c *Client) CreateStream(ctx context.Context, model CreateModel) (<-chan string, error) {
	model.Stream = true
	statusChan := make(chan string)
	return statusChan, nil
}
