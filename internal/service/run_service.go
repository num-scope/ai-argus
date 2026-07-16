package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/dao"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/internal/model"
)

func StartRun(ctx context.Context, req dto.StartRunRequest) (*dto.RunResponse, error) {
	target, err := dao.GetTargetByID(ctx, req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("读取目标: %w", err)
	}
	scenario, err := dao.GetScenarioByID(ctx, req.ScenarioID)
	if err != nil {
		return nil, fmt.Errorf("读取场景: %w", err)
	}
	runs.Lock()
	maxConcurrency := runs.maxConcurrency
	rootContext := runs.rootContext
	runs.Unlock()
	if scenario.Concurrency > maxConcurrency {
		return nil, invalidRequest("场景并发数超过当前平台上限")
	}
	cfg, err := buildBenchmarkConfig(*target, *scenario)
	if err != nil {
		return nil, err
	}
	snapshot := redactConfig(cfg)
	configJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	run := &model.Run{
		TargetID:     target.ID,
		ScenarioID:   scenario.ID,
		TargetName:   target.Name,
		ScenarioName: scenario.Name,
		Protocol:     target.Protocol,
		Model:        target.Model,
		Status:       model.RunStatusQueued,
		ConfigJSON:   string(configJSON),
		SummaryJSON:  "{}",
		Planned:      scenario.TotalRequests,
	}
	if err := dao.CreateRun(ctx, run); err != nil {
		return nil, err
	}

	runContext, cancel := context.WithCancel(rootContext)
	runs.Lock()
	runs.cancels[run.ID] = cancel
	runs.wait.Add(1)
	runs.Unlock()
	response := toRunResponse(*run)
	go executeRun(runContext, run, cfg)
	return &response, nil
}

func CancelRun(ctx context.Context, id int64) error {
	runs.Lock()
	cancel := runs.cancels[id]
	runs.Unlock()
	if cancel != nil {
		cancel()
		return nil
	}
	now := time.Now()
	cancelled, err := dao.CancelRunIfActive(ctx, id, now, "任务在服务重启后已失去执行上下文")
	if err != nil {
		return err
	}
	if cancelled {
		return nil
	}
	if _, err := dao.GetRunByID(ctx, id); err != nil {
		return err
	}
	return common.ErrRunAlreadyFinished
}

func ListRuns(ctx context.Context, limit int) ([]dto.RunResponse, error) {
	items, err := dao.ListRuns(ctx, limit)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.RunResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toRunResponse(item))
	}
	return responses, nil
}

func GetRunWorkspace(ctx context.Context, limit int) (*dto.RunWorkspace, error) {
	items, err := ListRuns(ctx, limit)
	if err != nil {
		return nil, err
	}
	targets, err := ListTargets(ctx)
	if err != nil {
		return nil, err
	}
	scenarios, err := ListScenarios(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.RunWorkspace{Runs: items, Targets: targets, Scenarios: scenarios}, nil
}

func GetRunDetail(ctx context.Context, id int64, resultLimit int) (*dto.RunDetail, error) {
	run, err := dao.GetRunByID(ctx, id)
	if err != nil {
		return nil, err
	}
	var results []model.RequestResult
	if resultLimit <= 0 {
		// Full chronological log for finished runs / PDF export.
		results, err = dao.ListRequestResultsASC(ctx, id)
	} else {
		// Newest-first bounded window for live HTMX refresh.
		results, err = dao.ListRequestResults(ctx, id, resultLimit)
	}
	if err != nil {
		return nil, err
	}
	var summary dto.RunSummary
	if run.SummaryJSON != "" {
		if err := json.Unmarshal([]byte(run.SummaryJSON), &summary); err != nil {
			return nil, fmt.Errorf("解析运行 %d 汇总: %w", run.ID, err)
		}
	}
	resultResponses := make([]dto.RequestResultResponse, 0, len(results))
	for _, result := range results {
		resultResponses = append(resultResponses, toRequestResultResponse(result))
	}
	return &dto.RunDetail{Run: toRunResponse(*run), Summary: summary, Results: resultResponses}, nil
}

// GetRunDetailForPage loads run detail with a result strategy suitable for the
// merged live+report UI: live runs keep a bounded newest window; finished runs
// load every request log (chronological) so on-screen view and PDF export match.
func GetRunDetailForPage(ctx context.Context, id int64) (*dto.RunDetail, error) {
	run, err := dao.GetRunByID(ctx, id)
	if err != nil {
		return nil, err
	}
	limit := 0
	switch run.Status {
	case model.RunStatusQueued, model.RunStatusWarming, model.RunStatusRunning:
		limit = 100
	}
	return GetRunDetail(ctx, id, limit)
}

// ExportRunJSONL returns newline-delimited JSON for every request result of a run.
func ExportRunJSONL(ctx context.Context, id int64) ([]byte, string, error) {
	run, err := dao.GetRunByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	results, err := dao.ListRequestResultsASC(ctx, id)
	if err != nil {
		return nil, "", err
	}
	var builder strings.Builder
	for _, result := range results {
		row := map[string]any{
			"run_id":            result.RunID,
			"request_index":     result.RequestIndex,
			"prompt":            result.Prompt,
			"ok":                result.OK,
			"status":            result.Status,
			"attempts":          result.Attempts,
			"elapsed_ms":        result.ElapsedMS,
			"queue_ms":          result.QueueMS,
			"request_ms":        result.RequestMS,
			"ttft_ms":           result.TTFTMS,
			"tpot_ms":           result.TPOTMS,
			"prompt_tokens":     result.PromptTokens,
			"completion_tokens": result.CompletionTokens,
			"total_tokens":      result.TotalTokens,
			"content_chunks":    result.ContentChunks,
			"streamed":          result.Streamed,
			"stream_completed":  result.StreamCompleted,
			"response_preview":  result.ResponsePreview,
			"error":             result.ErrorMessage,
			"created_at":        result.CreatedAt,
		}
		raw, marshalErr := json.Marshal(row)
		if marshalErr != nil {
			return nil, "", marshalErr
		}
		builder.Write(raw)
		builder.WriteByte('\n')
	}
	filename := fmt.Sprintf("run-%04d.jsonl", run.ID)
	return []byte(builder.String()), filename, nil
}

func ReconcileInterruptedRuns(ctx context.Context) error {
	return dao.FailInterruptedRuns(ctx)
}
