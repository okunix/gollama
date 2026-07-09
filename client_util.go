package gollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) parseError(resp *http.Response) error {
	var ollamaErr Error
	err := json.NewDecoder(resp.Body).Decode(&ollamaErr)
	if err != nil {
		return fmt.Errorf("ollama error status code: %d", resp.StatusCode)
	}
	return fmt.Errorf("ollama error response: %w", ollamaErr)
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

func (c *Client) toBody(a any) (io.Reader, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
