package gollama

import (
	"bufio"
	"context"
	"encoding/json"
	"iter"
	"net/http"
	"time"
)

type Client struct {
	token  string
	host   string
	client *http.Client
}

type ClientOption func(c *Client) error

// NewClient creates a new Client for interacting with an Ollama server,
// applying any given options to configure it. It verifies that the server
// is reachable before returning.
//
// It returns an error if any option fails to apply or the server cannot
// be reached.
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	client := Client{client: http.DefaultClient}
	for _, opt := range opts {
		if err := opt(&client); err != nil {
			return nil, err
		}
	}
	err := client.Ping(ctx)
	return &client, err
}

// WithToken configures the Client to authenticate requests using the
// given token.
func WithToken(token string) ClientOption {
	return func(c *Client) error {
		c.token = token
		return nil
	}
}

// WithHost configures the Client to send requests to the given Ollama
// server host.
func WithHost(host string) ClientOption {
	return func(c *Client) error {
		c.host = host
		return nil
	}
}

// WithTimeout configures the Client's underlying HTTP client to time out
// requests after the given duration.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) error {
		c.client.Timeout = timeout
		return nil
	}
}

// Ping checks whether the Ollama server is reachable and responding.
// It returns an error if the request fails or the server responds with a non-200 status code.
//
// A nil error indicates the server is up and reachable.
func (c *Client) Ping(ctx context.Context) error {
	req, err := c.newRequest(ctx, "GET", c.host, nil)
	if err != nil {
		return err
	}
	if _, err := c.do(req); err != nil {
		return err
	}
	return nil
}

// Version retrieves the version of the running Ollama server.
// It returns the version string reported by the server.
//
// It returns an error if the request fails, the server responds with
// a non-200 status code, or the response body cannot be decoded.
func (c *Client) Version(ctx context.Context) (string, error) {
	type response struct {
		Version string `json:"version"`
	}

	url := c.host + "/api/version"
	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	var resp response
	if err := c.decode(req, &resp); err != nil {
		return "", err
	}

	return resp.Version, nil
}

// Tags lists the models currently available locally on the Ollama server.
// It returns the list  of models, including details such as name, size, and modification time.
//
// It returns an error if the request fails, the server responds with
// a non-200 status code, or the response body cannot be decoded.
func (c *Client) Tags(ctx context.Context) ([]Model, error) {
	type response struct {
		Models []Model `json:"models"`
	}

	url := c.host + "/api/tags"
	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var resp response
	if err := c.decode(req, &resp); err != nil {
		return nil, err
	}

	return resp.Models, nil
}

// Ps lists the models currently loaded into memory on the Ollama server.
// It details about each running model
//
// It returns an error if the request fails, the server responds with
// a non-200 status code, or the response body cannot be decoded.
func (c *Client) Ps(ctx context.Context) ([]RunningModel, error) {
	type response struct {
		Models []RunningModel `json:"models"`
	}
	url := c.host + "/api/ps"
	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var resp response
	if err := c.decode(req, &resp); err != nil {
		return nil, err
	}
	return resp.Models, nil
}

// ShowModelDetails retrieves detailed information about a specific model,
// including its modelfile, template, parameters, and license.
//
// If verbose is true, the response includes additional detailed information
// such as full tokenizer data, rather than a truncated summary.
//
// It returns an error if the request fails, the server responds with
// a non-200 status code, or the response body cannot be decoded.
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

	var resp ModelDetails
	if err := c.decode(req, &resp); err != nil {
		return ModelDetails{}, err
	}
	return resp, nil
}

// Delete removes a model and its associated data from the Ollama server.
//
// It returns an error if the request fails or the server responds with
// a non-200 status code.
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

	if err := c.decode(req, nil); err != nil {
		return err
	}
	return nil
}

// Create creates a new model,
// It waits for the operation to complete before returning.
//
// It returns an error if the request fails or the server reports a failure
// while creating the model. Use CreateStream if you need progress updates
// as the model is being created.
func (c *Client) Create(ctx context.Context, model CreateRequest) error {
	model.Stream = false
	resp, err := c.create(ctx, model)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// CreateStream creates a new model and returns an
// iterator that yields status updates as the model is created
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
	return c.do(req)
}

// Copy duplicates an existing model under a new name.
//
// It returns an error if the request fails, the source model does not
// exist, or the server responds with a non-200 status code.
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
	return c.decode(req, nil)
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
	return c.do(req)
}

// Pull downloads a model to the local Ollama server, waiting for the
// download to complete before returning.
//
// If insecure is true, TLS verification is skipped when connecting to
// the model's registry; this should only be used for trusted registries.
//
// It returns an error if the model cannot be downloaded.
func (c *Client) Pull(ctx context.Context, model string, insecure bool) error {
	resp, err := c.pull(ctx, model, insecure, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// PullStream downloads a model to the local Ollama server and returns an
// iterator that yields status updates as the download progresses, allowing
// progress to be observed incrementally rather than waiting for the whole
// operation to complete.
//
// If insecure is true, TLS verification is skipped when connecting to
// the model's registry; this should only be used for trusted registries.
//
// It returns an error immediately if the download cannot be started.
// Iteration yields a non-nil error and stops if the download fails partway
// through.
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
	return c.do(req)
}

// Push uploads a model to a registry, waiting for the upload to complete
// before returning.
//
// If insecure is true, TLS verification is skipped when connecting to
// the registry; this should only be used for trusted registries.
//
// It returns an error if the model cannot be uploaded.
func (c *Client) Push(ctx context.Context, model string, insecure bool) error {
	resp, err := c.push(ctx, model, insecure, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// PushStream uploads a model to a registry and returns an iterator that
// yields status updates as the upload progresses, allowing progress to
// be observed incrementally rather than waiting for the whole operation
// to complete.
//
// If insecure is true, TLS verification is skipped when connecting to
// the registry; this should only be used for trusted registries.
//
// It returns an error immediately if the upload cannot be started.
// Iteration yields a non-nil error and stops if the upload fails partway
// through.
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
	return c.do(req)
}

// Generate produces a completion for the given prompt using a GenerateRequest,
// waiting for the full response before returning it.
//
// It returns an error if the completion cannot be generated or the response
// cannot be read. Use GenerateStream if you need the response incrementally
// as it is produced.
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

// GenerateStream produces a completion for the given prompt using a
// GenerateRequest and returns an iterator that yields response chunks as
// they are produced, allowing the completion to be consumed incrementally
// rather than waiting for it to finish in full.
//
// It returns an error immediately if the generation cannot be started.
// Iteration yields a non-nil error and stops if the generation fails
// partway through.
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

// Embed generates embeddings for the input given in an EmbedRequest,
// returning the resulting vectors along with metadata such as the model
// used and token counts.
//
// It returns an error if the embeddings cannot be generated.
func (c *Client) Embed(ctx context.Context, embedReq EmbedRequest) (EmbedResponse, error) {
	var embedResp EmbedResponse
	url := c.host + "/api/embed"
	body, _ := c.toBody(embedReq)
	req, err := c.newRequest(ctx, "POST", url, body)
	if err != nil {
		return embedResp, err
	}
	err = c.decode(req, &embedResp)
	return embedResp, err
}

func (c *Client) chat(ctx context.Context, chatReq ChatRequest) (*http.Response, error) {
	body, _ := c.toBody(chatReq)
	url := c.host + "/api/chat"
	req, err := c.newRequest(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// Chat generates the next message in a conversation using a ChatRequest,
// waiting for the full response before returning it.
//
// It returns an error if the response cannot be generated or the response
// cannot be read. Use ChatStream if you need the response incrementally
// as it is produced.
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

// ChatStream generates the next message in a conversation using a
// ChatRequest and returns an iterator that yields response chunks as
// they are produced, allowing the response to be consumed incrementally
// rather than waiting for it to finish in full.
//
// It returns an error immediately if the chat cannot be started.
// Iteration yields a non-nil error and stops if generation fails partway
// through.
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
