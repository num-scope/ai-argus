package model

import "time"

const (
	RunStatusQueued    = "queued"
	RunStatusWarming   = "warming"
	RunStatusRunning   = "running"
	RunStatusCompleted = "completed"
	RunStatusFailed    = "failed"
	RunStatusCancelled = "cancelled"
)

type Run struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	TargetID     int64      `gorm:"index;not null" json:"target_id"`
	ScenarioID   int64      `gorm:"index;not null" json:"scenario_id"`
	TargetName   string     `gorm:"size:120;not null" json:"target_name"`
	ScenarioName string     `gorm:"size:120;not null" json:"scenario_name"`
	Protocol     string     `gorm:"size:24;not null" json:"protocol"`
	Model        string     `gorm:"size:200;not null" json:"model"`
	Status       string     `gorm:"size:24;index;not null" json:"status"`
	ConfigJSON   string     `gorm:"type:text;not null" json:"-"`
	SummaryJSON  string     `gorm:"type:text" json:"-"`
	Planned      int        `gorm:"not null" json:"planned"`
	Completed    int        `gorm:"not null" json:"completed"`
	Succeeded    int        `gorm:"not null" json:"succeeded"`
	Failed       int        `gorm:"not null" json:"failed"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RequestResult struct {
	ID               int64     `gorm:"primaryKey" json:"id"`
	RunID            int64     `gorm:"index;not null" json:"run_id"`
	RequestIndex     int       `gorm:"not null" json:"request_index"`
	Prompt           string    `gorm:"type:text;not null" json:"prompt"`
	OK               bool      `gorm:"index;not null" json:"ok"`
	Status           *int      `json:"status,omitempty"`
	Attempts         int       `gorm:"not null" json:"attempts"`
	ElapsedMS        float64   `gorm:"not null" json:"elapsed_ms"`
	QueueMS          float64   `gorm:"not null" json:"queue_ms"`
	RequestMS        float64   `gorm:"not null" json:"request_ms"`
	TTFTMS           *float64  `json:"ttft_ms,omitempty"`
	TPOTMS           *float64  `json:"tpot_ms,omitempty"`
	PromptTokens     *int      `json:"prompt_tokens,omitempty"`
	CompletionTokens *int      `json:"completion_tokens,omitempty"`
	TotalTokens      *int      `json:"total_tokens,omitempty"`
	ContentChunks    int       `gorm:"not null" json:"content_chunks"`
	Streamed         bool      `gorm:"not null" json:"streamed"`
	StreamCompleted  bool      `gorm:"not null" json:"stream_completed"`
	ResponsePreview  string    `gorm:"type:text" json:"response_preview,omitempty"`
	ErrorMessage     string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
