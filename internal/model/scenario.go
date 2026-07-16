package model

import "time"

type Scenario struct {
	ID                      int64     `gorm:"primaryKey" json:"id"`
	Name                    string    `gorm:"size:120;not null;uniqueIndex" json:"name"`
	SystemPrompt            string    `gorm:"type:text" json:"system_prompt"`
	PromptsJSON             string    `gorm:"type:text;not null" json:"-"`
	Concurrency             int       `gorm:"not null" json:"concurrency"`
	TotalRequests           int       `gorm:"not null" json:"total_requests"`
	WarmupRequests          int       `gorm:"not null" json:"warmup_requests"`
	RampUpSeconds           float64   `gorm:"not null" json:"ramp_up_seconds"`
	Temperature             float64   `gorm:"not null" json:"temperature"`
	TopP                    float64   `gorm:"not null" json:"top_p"`
	MaxOutputTokens         int       `gorm:"not null" json:"max_output_tokens"`
	Seed                    *int      `json:"seed"`
	IncludeUsage            bool      `gorm:"not null" json:"include_usage"`
	TimeoutSeconds          float64   `gorm:"not null" json:"timeout_seconds"`
	ConnectTimeoutSeconds   float64   `gorm:"column:connect_timeout_secs;not null" json:"connect_timeout_seconds"`
	MaxRetries              int       `gorm:"not null" json:"max_retries"`
	RetryBaseDelaySeconds   float64   `gorm:"column:retry_base_delay_secs;not null" json:"retry_base_delay_seconds"`
	SaveResponsePreview     bool      `gorm:"not null" json:"save_response_preview"`
	SaveResponseBody        bool      `gorm:"not null;default:false" json:"save_response_body"`
	ResponsePreviewLength   int       `gorm:"not null;default:100" json:"response_preview_length"`
	RandomPromptMode        bool      `gorm:"not null;default:false" json:"random_prompt_mode"`
	RandomPromptTargetChars int       `gorm:"not null;default:20" json:"random_prompt_target_chars"`
	RandomPromptMaxChars    int       `gorm:"not null;default:28" json:"random_prompt_max_chars"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}
