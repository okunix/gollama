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

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	// ollama uses 200 status code as a general success
	// any other codes are considered to be error related
	// https://docs.ollama.com/api/errors#status-codes
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, c.parseError(resp)
	}
	return resp, nil
}

func (c *Client) decode(req *http.Request, dest any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if dest == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}
