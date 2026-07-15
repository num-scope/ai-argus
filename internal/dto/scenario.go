package dto

import "time"

type ScenarioRequest struct {
	Name                  string  `form:"name" json:"name"`
	SystemPrompt          string  `form:"system_prompt" json:"system_prompt"`
	Prompts               string  `form:"prompts" json:"prompts"`
	Concurrency           int     `form:"concurrency" json:"concurrency"`
	TotalRequests         int     `form:"total_requests" json:"total_requests"`
	WarmupRequests        int     `form:"warmup_requests" json:"warmup_requests"`
	RampUpSeconds         float64 `form:"ramp_up_seconds" json:"ramp_up_seconds"`
	Temperature           float64 `form:"temperature" json:"temperature"`
	TopP                  float64 `form:"top_p" json:"top_p"`
	MaxOutputTokens       int     `form:"max_output_tokens" json:"max_output_tokens"`
	Seed                  string  `form:"seed" json:"seed"`
	IncludeUsage          bool    `form:"include_usage" json:"include_usage"`
	TimeoutSeconds        float64 `form:"timeout_seconds" json:"timeout_seconds"`
	ConnectTimeoutSeconds float64 `form:"connect_timeout_seconds" json:"connect_timeout_seconds"`
	MaxRetries            int     `form:"max_retries" json:"max_retries"`
	RetryBaseDelaySeconds float64 `form:"retry_base_delay_seconds" json:"retry_base_delay_seconds"`
	SaveResponsePreview   bool    `form:"save_response_preview" json:"save_response_preview"`
}

type ScenarioResponse struct {
	ID                    int64     `json:"id"`
	Name                  string    `json:"name"`
	SystemPrompt          string    `json:"system_prompt"`
	Prompts               []string  `json:"prompts"`
	Concurrency           int       `json:"concurrency"`
	TotalRequests         int       `json:"total_requests"`
	WarmupRequests        int       `json:"warmup_requests"`
	RampUpSeconds         float64   `json:"ramp_up_seconds"`
	Temperature           float64   `json:"temperature"`
	TopP                  float64   `json:"top_p"`
	MaxOutputTokens       int       `json:"max_output_tokens"`
	Seed                  *int      `json:"seed,omitempty"`
	IncludeUsage          bool      `json:"include_usage"`
	TimeoutSeconds        float64   `json:"timeout_seconds"`
	ConnectTimeoutSeconds float64   `json:"connect_timeout_seconds"`
	MaxRetries            int       `json:"max_retries"`
	RetryBaseDelaySeconds float64   `json:"retry_base_delay_seconds"`
	SaveResponsePreview   bool      `json:"save_response_preview"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
