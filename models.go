package gollama

import (
	"fmt"
	"strings"
	"time"
)

type Error struct {
	Err string `json:"error"`
}

func (e Error) Error() string {
	return e.Err
}

type Model struct {
	Name        string      `json:"name"`
	Model       string      `json:"model"`
	RemoteModel string      `json:"remote_model"`
	RemoteHost  string      `json:"remote_host"`
	ModifiedAt  time.Time   `json:"modified_at"`
	Size        int64       `json:"size"`
	Digest      string      `json:"digest"`
	Details     ModelDetail `json:"details"`
}

type ModelDetail struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type Ps struct {
	Name          string      `json:"name"`
	Model         string      `json:"model"`
	Size          int64       `json:"size"`
	Digest        string      `json:"digest"`
	Details       ModelDetail `json:"details"`
	ExpiresAt     time.Time   `json:"expires_at"`
	SizeVRAM      int64       `json:"size_vram"`
	ContextLength int64       `json:"context_length"`
}

type ModelDetails struct {
	Parameters   string         `json:"parameters"`
	License      string         `json:"license"`
	ModifiedAt   time.Time      `json:"modified_at"`
	Details      ModelDetail    `json:"details"`
	Template     string         `json:"template"`
	Capabilities []string       `json:"capabilities"`
	ModelInfo    map[string]any `json:"model_info"`
}

type CreateRequest struct {
	Model      string         `json:"model"`
	From       string         `json:"from"`
	Template   string         `json:"template"`
	License    string         `json:"license"`
	System     string         `json:"system"`
	Parameters map[string]any `json:"parameters"`
	Messages   []Message      `json:"messages"`
	Quantize   string         `json:"quantize"`
	Stream     bool           `json:"stream"`
}

func (c CreateRequest) Validate() error {
	if len(strings.TrimSpace(c.Model)) == 0 {
		return fmt.Errorf("model name is required")
	}
	return nil
}

type Role string

const (
	SYSTEM    Role = "system"
	USER      Role = "user"
	ASSISTANT Role = "assistant"
	TOOL      Role = "tool"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	// each image must encoded with base64
	Images    []string   `json:"images"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Arguments   map[string]any `json:"arguments"`
}

type Status struct {
	Status    string  `json:"status"`
	Digest    string  `json:"digest"`
	Total     int64   `json:"total"`
	Completed int64   `json:"completed"`
	Error     *string `json:"error,omitempty"`
}

type GenerateRequest struct {
	Model       string           `json:"model"`
	Prompt      string           `json:"prompt,omitempty"`
	Suffix      string           `json:"suffix,omitempty"`
	Images      []string         `json:"images,omitempty"`
	Format      string           `json:"format,omitempty"`
	System      string           `json:"system,omitempty"`
	Stream      bool             `json:"stream"`
	Think       bool             `json:"bool,omitempty"`
	Raw         bool             `json:"raw,omitempty"`
	KeepAlive   string           `json:"keep_alive,omitempty"`
	Options     []GenerateOption `json:"options,omitempty"`
	Logprobs    bool             `json:"lobprobs,omitempty"`
	TopLogprobs int64            `json:"top_logprobs,omitempty"`
}

type GenerateOption struct {
	Seed        int64    `json:"seed,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	TopK        int64    `json:"top_k,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	MinP        float64  `json:"min_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	NumCtx      int64    `json:"num_ctx,omitempty"`
	NumPredict  int64    `json:"num_predict,omitempty"`
}

type GenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Thinking           string    `json:"thinking"`
	Done               bool      `json:"done"`
	DoneReason         string    `json:"done_reason"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int64     `json:"load_duration"`
	PromptEvalCount    int64     `json:"prompt_eval_count"`
	PromptEvalDuration int64     `json:"prompt_eval_duration"`
	EvalCount          int64     `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
	Logprobs           []Logprob `json:"logprobs"`
}

type Logprob struct {
	Token       string       `json:"token"`
	Logprob     int64        `json:"logprob"`
	Bytes       []int64      `json:"bytes"`
	TopLogprobs []TopLogprob `json:"top_logprobs"`
}

type TopLogprob struct {
	Token   string  `json:"token"`
	Logprob int64   `json:"logprob"`
	Bytes   []int64 `json:"bytes"`
}

type GenerateStreamResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Thinking           string    `json:"thinking"`
	Done               bool      `json:"done"`
	DoneReason         string    `json:"done_reason"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int64     `json:"load_duration"`
	PromptEvalCount    int64     `json:"prompt_eval_count"`
	PromptEvalDuration int64     `json:"prompt_eval_duration"`
	EvalCount          int64     `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
	Error              *string   `json:"error,omitempty"`
}

type EmbedRequest struct {
	Model      string    `json:"model"`
	Input      string    `json:"input"`
	Truncate   bool      `json:"truncate"`
	Dimensions int64     `json:"dimensions,omitempty"`
	KeepAlive  string    `json:"keep_alive,omitempty"`
	Options    []Options `json:"options"`
}

type Options struct {
	Seed        int64   `json:"seed"`
	Temperature float64 `json:"temperature"`
	TopK        int64   `json:"top_k"`
	TopP        float64 `json:"top_p"`
	MinP        float64 `json:"min_p"`
	Stop        string  `json:"stop"`
	NumCtx      int64   `json:"num_ctx"`
	NumPredict  int64   `json:"num_predict"`
}

type EmbedResponse struct {
	Model           string      `json:"model"`
	Embeddings      [][]float64 `json:"embeddings"`
	TotalDuration   int64       `json:"total_duration"`
	LoadDuration    int64       `json:"load_duration"`
	PromptEvalCount int64       `json:"prompt_eval_count"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Format      string    `json:"format,omitempty"`
	Options     []Options `json:"options,omitempty"`
	Tools       []Tool    `json:"tools"`
	Stream      bool      `json:"stream"`
	Think       bool      `json:"think,omitempty"`
	KeepAlive   string    `json:"keep_alive,omitempty"`
	Logprobs    bool      `json:"logprobs"`
	TopLogprobs int64     `json:"top_logprobs"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string `json:"name"`
	Parameters  string `json:"parameters"`
	Description string `json:"description,omitempty"`
}

type ChatResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	DoneReason         string    `json:"done_reason"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int64     `json:"load_duration"`
	PromptEvalCount    int64     `json:"prompt_eval_count"`
	PromptEvalDuration int64     `json:"prompt_eval_duration"`
	EvalCount          int64     `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
	Logprobs           []Logprob `json:"logprobs"`
}

type ChatStreamResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Message   Message   `json:"message"`
	Done      bool      `json:"done"`
	Error     *string   `json:"error,omitempty"`
}
