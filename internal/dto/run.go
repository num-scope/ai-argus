package dto

import "time"

type StartRunRequest struct {
	TargetID   int64 `form:"target_id" json:"target_id"`
	ScenarioID int64 `form:"scenario_id" json:"scenario_id"`
}

type PercentileMetric struct {
	Average float64 `json:"average"`
	P50     float64 `json:"p50"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
	Samples int     `json:"samples"`
}

type RunSummary struct {
	DurationSeconds        float64          `json:"duration_seconds"`
	RPS                    float64          `json:"rps"`
	SuccessRate            float64          `json:"success_rate"`
	Retries                int              `json:"retries"`
	CompleteResponses      int              `json:"complete_responses"`
	StatusCounts           map[string]int   `json:"status_counts"`
	PromptTokens           int              `json:"prompt_tokens"`
	CompletionTokens       int              `json:"completion_tokens"`
	TotalTokens            int              `json:"total_tokens"`
	UsageCoverage          float64          `json:"usage_coverage"`
	GenerationTokensPerSec float64          `json:"generation_tokens_per_second"`
	CompletionTokensPerSec float64          `json:"completion_tokens_per_second"`
	E2E                    PercentileMetric `json:"e2e_ms"`
	Queue                  PercentileMetric `json:"queue_ms"`
	Request                PercentileMetric `json:"request_ms"`
	TTFT                   PercentileMetric `json:"ttft_ms"`
	TPOT                   PercentileMetric `json:"tpot_ms"`
}

type RunResponse struct {
	ID           int64      `json:"id"`
	TargetID     int64      `json:"target_id"`
	ScenarioID   int64      `json:"scenario_id"`
	TargetName   string     `json:"target_name"`
	ScenarioName string     `json:"scenario_name"`
	Protocol     string     `json:"protocol"`
	Model        string     `json:"model"`
	Status       string     `json:"status"`
	Planned      int        `json:"planned"`
	Completed    int        `json:"completed"`
	Succeeded    int        `json:"succeeded"`
	Failed       int        `json:"failed"`
	ErrorMessage string     `json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RequestResultResponse struct {
	ID               int64     `json:"id"`
	RunID            int64     `json:"run_id"`
	RequestIndex     int       `json:"request_index"`
	Prompt           string    `json:"prompt"`
	OK               bool      `json:"ok"`
	Status           *int      `json:"status,omitempty"`
	Attempts         int       `json:"attempts"`
	ElapsedMS        float64   `json:"elapsed_ms"`
	QueueMS          float64   `json:"queue_ms"`
	RequestMS        float64   `json:"request_ms"`
	TTFTMS           *float64  `json:"ttft_ms,omitempty"`
	TPOTMS           *float64  `json:"tpot_ms,omitempty"`
	PromptTokens     *int      `json:"prompt_tokens,omitempty"`
	CompletionTokens *int      `json:"completion_tokens,omitempty"`
	TotalTokens      *int      `json:"total_tokens,omitempty"`
	ContentChunks    int       `json:"content_chunks"`
	Streamed         bool      `json:"streamed"`
	StreamCompleted  bool      `json:"stream_completed"`
	ResponsePreview  string    `json:"response_preview,omitempty"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type RunDetail struct {
	Run     RunResponse             `json:"run"`
	Summary RunSummary              `json:"summary"`
	Results []RequestResultResponse `json:"results"`
}

type RunWorkspace struct {
	Runs      []RunResponse      `json:"runs"`
	Targets   []TargetResponse   `json:"targets"`
	Scenarios []ScenarioResponse `json:"scenarios"`
}

type Dashboard struct {
	TargetCount     int64
	ScenarioCount   int64
	RunCount        int64
	ActiveRunCount  int64
	RecentRuns      []RunResponse
	LastSuccessRate float64
	LastRPS         float64
	UpdatedAt       time.Time
}
