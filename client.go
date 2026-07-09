package gollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
	defer resp.Body.Close()

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

func (c *Client) Ps(ctx context.Context) ([]RunningModel, error) {
	type response struct {
		Models []RunningModel `json:"models"`
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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
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

func (c *Client) Create(ctx context.Context, model CreateRequest) error {
	model.Stream = false
	_, err := c.create(ctx, model)
	return err
}

func (c *Client) CreateStream(
	ctx context.Context,
	model CreateRequest,
) (iter.Seq2[Status, error], error) {
	model.Stream = true
	resp, err := c.create(ctx, model)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(resp.Body)
	return func(yield func(Status, error) bool) {
		defer resp.Body.Close()
		for scanner.Scan() {
			line := scanner.Text()
			var status Status

			err := json.Unmarshal([]byte(line), &status)
			if err != nil {
				yield(Status{}, err)
				return
			}
			if status.Error != nil {
				yield(Status{}, Error{Err: *status.Error})
				return
			}

			if !yield(status, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield(Status{}, err)
		}
	}, nil
}

func (c *Client) create(ctx context.Context, model CreateRequest) (*http.Response, error) {
	url := c.host + "/api/create"
	body, _ := c.toBody(model)
	req, err := c.newRequest(ctx, "POST", url, body)
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
	return resp, nil
}

func (c *Client) Copy(ctx context.Context, src, dest string) error {
	type request struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	url := c.host + "/api/copy"
	body, _ := c.toBody(request{Source: src, Destination: dest})
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

func (c *Client) pull(
	ctx context.Context,
	model string,
	insecure, stream bool,
) (*http.Response, error) {
	type request struct {
		Model    string `json:"model"`
		Insecure bool   `json:"insecure"`
		Stream   bool   `json:"stream"`
	}
	url := c.host + "/api/pull"
	requestModel := request{Model: model, Insecure: insecure, Stream: stream}
	body, _ := c.toBody(requestModel)
	req, err := c.newRequest(ctx, "POST", url, body)
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
	return resp, nil
}

func (c *Client) Pull(ctx context.Context, model string, insecure bool) error {
	_, err := c.pull(ctx, model, insecure, false)
	return err
}

func (c *Client) PullStream(
	ctx context.Context,
	model string,
	insecure bool,
) (iter.Seq2[Status, error], error) {
	resp, err := c.pull(ctx, model, insecure, true)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(resp.Body)
	return func(yield func(Status, error) bool) {
		defer resp.Body.Close()
		for scanner.Scan() {
			line := scanner.Text()

			var status Status

			err := json.Unmarshal([]byte(line), &status)
			if err != nil {
				yield(Status{}, err)
				return
			}
			if status.Error != nil {
				yield(Status{}, Error{Err: *status.Error})
				return
			}

			if !yield(status, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield(Status{}, err)
			return
		}
	}, nil
}

func (c *Client) push(
	ctx context.Context,
	model string,
	insecure, stream bool,
) (*http.Response, error) {
	type request struct {
		Model    string `json:"model"`
		Insecure bool   `json:"insecure"`
		Stream   bool   `json:"stream"`
	}
	url := c.host + "/api/push"
	requestModel := request{Model: model, Insecure: insecure, Stream: stream}
	body, _ := c.toBody(requestModel)
	req, err := c.newRequest(ctx, "POST", url, body)
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
	return resp, nil
}

func (c *Client) Push(ctx context.Context, model string, insecure bool) error {
	_, err := c.push(ctx, model, insecure, false)
	return err
}

func (c *Client) PushStream(
	ctx context.Context,
	model string,
	insecure bool,
) (iter.Seq2[Status, error], error) {
	resp, err := c.push(ctx, model, insecure, true)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(resp.Body)
	return func(yield func(Status, error) bool) {
		defer resp.Body.Close()
		for scanner.Scan() {
			line := scanner.Text()

			var status Status

			err := json.Unmarshal([]byte(line), &status)
			if err != nil {
				yield(Status{}, err)
				return
			}
			if status.Error != nil {
				yield(Status{}, Error{Err: *status.Error})
				return
			}

			if !yield(status, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield(Status{}, err)
			return
		}
	}, nil
}

func (c *Client) generate(ctx context.Context, genReq GenerateRequest) (*http.Response, error) {
	body, _ := c.toBody(genReq)
	url := c.host + "/api/generate"
	req, err := c.newRequest(ctx, "POST", url, body)
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
	return resp, nil
}

func (c *Client) Generate(ctx context.Context, genReq GenerateRequest) (GenerateResponse, error) {
	genReq.Stream = false
	var genResp GenerateResponse
	resp, err := c.generate(ctx, genReq)
	if err != nil {
		return genResp, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&genResp)
	return genResp, err
}

func (c *Client) GenerateStream(
	ctx context.Context,
	genReq GenerateRequest,
) (iter.Seq2[GenerateStreamResponse, error], error) {
	genReq.Stream = true
	resp, err := c.generate(ctx, genReq)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(resp.Body)
	return func(yield func(GenerateStreamResponse, error) bool) {
		defer resp.Body.Close()
		for scanner.Scan() {
			line := scanner.Text()
			var genResp GenerateStreamResponse

			err := json.Unmarshal([]byte(line), &genResp)
			if err != nil {
				yield(GenerateStreamResponse{}, err)
				return
			}
			if genResp.Error != nil {
				yield(GenerateStreamResponse{}, Error{Err: *genResp.Error})
				return
			}

			if !yield(genResp, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield(GenerateStreamResponse{}, err)
		}
	}, nil
}

func (c *Client) Embed(ctx context.Context, embedReq EmbedRequest) (EmbedResponse, error) {
	var embedResp EmbedResponse
	url := c.host + "/api/embed"
	body, _ := c.toBody(embedReq)
	req, err := c.newRequest(ctx, "POST", url, body)
	if err != nil {
		return embedResp, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return embedResp, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return embedResp, c.parseError(resp)
	}
	err = json.NewDecoder(resp.Body).Decode(&embedResp)
	return embedResp, err
}

func (c *Client) chat(ctx context.Context, chatReq ChatRequest) (*http.Response, error) {
	body, _ := c.toBody(chatReq)
	url := c.host + "/api/chat"
	req, err := c.newRequest(ctx, "POST", url, body)
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
	return resp, nil
}

func (c *Client) Chat(ctx context.Context, chatReq ChatRequest) (ChatResponse, error) {
	chatReq.Stream = false
	var chatResp ChatResponse
	resp, err := c.chat(ctx, chatReq)
	if err != nil {
		return chatResp, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	return chatResp, err
}

func (c *Client) ChatStream(
	ctx context.Context,
	chatReq ChatRequest,
) (iter.Seq2[ChatStreamResponse, error], error) {
	chatReq.Stream = true
	resp, err := c.chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(resp.Body)
	return func(yield func(ChatStreamResponse, error) bool) {
		defer resp.Body.Close()
		for scanner.Scan() {
			var streamResp ChatStreamResponse
			line := scanner.Text()

			err := json.Unmarshal([]byte(line), &streamResp)
			if err != nil {
				yield(ChatStreamResponse{}, err)
				return
			}
			if streamResp.Error != nil {
				yield(ChatStreamResponse{}, err)
				return
			}

			if !yield(streamResp, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield(ChatStreamResponse{}, err)
		}
	}, nil
}
