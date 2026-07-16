package dto

import "time"

// Default scenario values aligned with ai-scripts practical defaults.
const (
	DefaultConcurrency             = 500
	DefaultTotalRequests           = 100
	DefaultWarmupRequests          = 0
	DefaultRampUpSeconds           = 0
	DefaultTemperature             = 0.7
	DefaultTopP                    = 1.0
	// DefaultMaxOutputTokens is 0: do not send max_tokens (provider default / unlimited).
	DefaultMaxOutputTokens         = 0
	DefaultTimeoutSeconds          = 0
	DefaultConnectTimeoutSeconds   = 10
	DefaultMaxRetries              = 2
	DefaultRetryBaseDelaySeconds   = 1.5
	DefaultResponsePreviewLength   = 100
	DefaultRandomPromptTargetChars = 20
	DefaultRandomPromptMaxChars    = 28
	MaxResponsePreviewLength       = 5000
	MaxStoredResponseBodyChars     = 32000
)

type ScenarioRequest struct {
	Name                    string  `form:"name" json:"name"`
	SystemPrompt            string  `form:"system_prompt" json:"system_prompt"`
	Prompts                 string  `form:"prompts" json:"prompts"`
	Concurrency             int     `form:"concurrency" json:"concurrency"`
	TotalRequests           int     `form:"total_requests" json:"total_requests"`
	WarmupRequests          int     `form:"warmup_requests" json:"warmup_requests"`
	RampUpSeconds           float64 `form:"ramp_up_seconds" json:"ramp_up_seconds"`
	Temperature             float64 `form:"temperature" json:"temperature"`
	TopP                    float64 `form:"top_p" json:"top_p"`
	MaxOutputTokens         int     `form:"max_output_tokens" json:"max_output_tokens"`
	Seed                    string  `form:"seed" json:"seed"`
	IncludeUsage            bool    `form:"include_usage" json:"include_usage"`
	TimeoutSeconds          float64 `form:"timeout_seconds" json:"timeout_seconds"`
	ConnectTimeoutSeconds   float64 `form:"connect_timeout_seconds" json:"connect_timeout_seconds"`
	MaxRetries              int     `form:"max_retries" json:"max_retries"`
	RetryBaseDelaySeconds   float64 `form:"retry_base_delay_seconds" json:"retry_base_delay_seconds"`
	SaveResponsePreview     bool    `form:"save_response_preview" json:"save_response_preview"`
	SaveResponseBody        bool    `form:"save_response_body" json:"save_response_body"`
	ResponsePreviewLength   int     `form:"response_preview_length" json:"response_preview_length"`
	RandomPromptMode        bool    `form:"random_prompt_mode" json:"random_prompt_mode"`
	RandomPromptTargetChars int     `form:"random_prompt_target_chars" json:"random_prompt_target_chars"`
	RandomPromptMaxChars    int     `form:"random_prompt_max_chars" json:"random_prompt_max_chars"`
}

type ScenarioResponse struct {
	ID                      int64     `json:"id"`
	Name                    string    `json:"name"`
	SystemPrompt            string    `json:"system_prompt"`
	Prompts                 []string  `json:"prompts"`
	Concurrency             int       `json:"concurrency"`
	TotalRequests           int       `json:"total_requests"`
	WarmupRequests          int       `json:"warmup_requests"`
	RampUpSeconds           float64   `json:"ramp_up_seconds"`
	Temperature             float64   `json:"temperature"`
	TopP                    float64   `json:"top_p"`
	MaxOutputTokens         int       `json:"max_output_tokens"`
	Seed                    *int      `json:"seed,omitempty"`
	IncludeUsage            bool      `json:"include_usage"`
	TimeoutSeconds          float64   `json:"timeout_seconds"`
	ConnectTimeoutSeconds   float64   `json:"connect_timeout_seconds"`
	MaxRetries              int       `json:"max_retries"`
	RetryBaseDelaySeconds   float64   `json:"retry_base_delay_seconds"`
	SaveResponsePreview     bool      `json:"save_response_preview"`
	SaveResponseBody        bool      `json:"save_response_body"`
	ResponsePreviewLength   int       `json:"response_preview_length"`
	RandomPromptMode        bool      `json:"random_prompt_mode"`
	RandomPromptTargetChars int       `json:"random_prompt_target_chars"`
	RandomPromptMaxChars    int       `json:"random_prompt_max_chars"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}
