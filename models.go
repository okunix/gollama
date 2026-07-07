package gollama

import "time"

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
