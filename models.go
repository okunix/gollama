package gollama

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// type Metrics struct {
// 	TotalDuration      int64 `json:"total_duration"`
// 	LoadDuration       int64 `json:"load_duration"`
// 	PromptEvalCount    int64 `json:"prompt_eval_count"`
// 	PromptEvalDuration int64 `json:"prompt_eval_duration"`
// 	EvalCount          int64 `json:"eval_count"`
// 	EvalDuration       int64 `json:"eval_duration"`
// }

type Error struct {
	Err string `json:"error"`
}

func (e Error) Error() string {
	return e.Err
}

func parseError(line string) (Error, error) {
	var ollamaError Error
	err := json.Unmarshal([]byte(line), &ollamaError)
	return ollamaError, err
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

type CreateModel struct {
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

func (c CreateModel) Validate() error {
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
