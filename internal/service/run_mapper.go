package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xtj/ai-argus/internal/benchmark"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/internal/model"
	"github.com/xtj/ai-argus/internal/protocol"
)

func buildBenchmarkConfig(target model.Target, scenario model.Scenario) (benchmark.Config, error) {
	prompts, err := decodePrompts(scenario.PromptsJSON)
	if err != nil {
		return benchmark.Config{}, fmt.Errorf("解析提示词: %w", err)
	}
	extraHeaders := make(map[string]string)
	if err := json.Unmarshal([]byte(defaultJSONObject(target.ExtraHeadersJSON)), &extraHeaders); err != nil {
		return benchmark.Config{}, fmt.Errorf("解析额外请求头: %w", err)
	}
	extraBody := make(map[string]any)
	if err := json.Unmarshal([]byte(defaultJSONObject(target.ExtraBodyJSON)), &extraBody); err != nil {
		return benchmark.Config{}, fmt.Errorf("解析额外请求体: %w", err)
	}
	previewLen := normalizePreviewLength(scenario.ResponsePreviewLength)
	targetChars := normalizeRandomTarget(scenario.RandomPromptTargetChars)
	maxChars := normalizeRandomMax(scenario.RandomPromptMaxChars, scenario.RandomPromptTargetChars)
	return benchmark.Config{
		Protocol: protocol.Config{
			URL:                   target.URL,
			APIKey:                target.APIKey,
			Protocol:              target.Protocol,
			Model:                 target.Model,
			CustomContentField:    target.CustomContentField,
			ExtraHeaders:          extraHeaders,
			ExtraBody:             extraBody,
			SystemPrompt:          scenario.SystemPrompt,
			Temperature:           scenario.Temperature,
			TopP:                  scenario.TopP,
			MaxOutputTokens:       scenario.MaxOutputTokens,
			Seed:                  scenario.Seed,
			IncludeUsage:          scenario.IncludeUsage,
			TimeoutSeconds:        scenario.TimeoutSeconds,
			ConnectTimeoutSeconds: scenario.ConnectTimeoutSeconds,
			MaxRetries:            scenario.MaxRetries,
			RetryBaseDelaySeconds: scenario.RetryBaseDelaySeconds,
			SaveResponsePreview:   scenario.SaveResponsePreview,
			SaveResponseBody:      scenario.SaveResponseBody,
			ResponsePreviewLength: previewLen,
		},
		Prompts:                 prompts,
		Concurrency:             scenario.Concurrency,
		TotalRequests:           scenario.TotalRequests,
		WarmupRequests:          scenario.WarmupRequests,
		RampUpSeconds:           scenario.RampUpSeconds,
		RandomPromptMode:        scenario.RandomPromptMode,
		RandomPromptTargetChars: targetChars,
		RandomPromptMaxChars:    maxChars,
	}, nil
}

func redactConfig(cfg benchmark.Config) benchmark.Config {
	cfg.Protocol.APIKey = ""
	headers := make(map[string]string, len(cfg.Protocol.ExtraHeaders))
	for key, value := range cfg.Protocol.ExtraHeaders {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "authorization") || strings.Contains(lower, "api-key") || strings.Contains(lower, "token") {
			headers[key] = "***"
		} else {
			headers[key] = value
		}
	}
	cfg.Protocol.ExtraHeaders = headers
	return cfg
}

func toRequestResult(runID int64, result protocol.Result) model.RequestResult {
	return model.RequestResult{
		RunID:            runID,
		RequestIndex:     result.Index,
		Prompt:           result.Prompt,
		OK:               result.OK,
		Status:           result.Status,
		Attempts:         result.Attempts,
		ElapsedMS:        result.ElapsedMS,
		QueueMS:          result.QueueMS,
		RequestMS:        result.RequestMS,
		TTFTMS:           result.TTFTMS,
		TPOTMS:           result.TPOTMS,
		PromptTokens:     result.PromptTokens,
		CompletionTokens: result.CompletionTokens,
		TotalTokens:      result.TotalTokens,
		ContentChunks:    result.ContentChunks,
		Streamed:         result.Streamed,
		StreamCompleted:  result.StreamCompleted,
		ResponsePreview:  result.ResponsePreview,
		ErrorMessage:     result.Error,
	}
}

func toRunResponse(run model.Run) dto.RunResponse {
	return dto.RunResponse{
		ID:           run.ID,
		TargetID:     run.TargetID,
		ScenarioID:   run.ScenarioID,
		TargetName:   run.TargetName,
		ScenarioName: run.ScenarioName,
		Protocol:     run.Protocol,
		Model:        run.Model,
		Status:       run.Status,
		Planned:      run.Planned,
		Completed:    run.Completed,
		Succeeded:    run.Succeeded,
		Failed:       run.Failed,
		ErrorMessage: run.ErrorMessage,
		StartedAt:    run.StartedAt,
		FinishedAt:   run.FinishedAt,
		CreatedAt:    run.CreatedAt,
		UpdatedAt:    run.UpdatedAt,
	}
}

func toRequestResultResponse(result model.RequestResult) dto.RequestResultResponse {
	return dto.RequestResultResponse{
		ID:               result.ID,
		RunID:            result.RunID,
		RequestIndex:     result.RequestIndex,
		Prompt:           result.Prompt,
		OK:               result.OK,
		Status:           result.Status,
		Attempts:         result.Attempts,
		ElapsedMS:        result.ElapsedMS,
		QueueMS:          result.QueueMS,
		RequestMS:        result.RequestMS,
		TTFTMS:           result.TTFTMS,
		TPOTMS:           result.TPOTMS,
		PromptTokens:     result.PromptTokens,
		CompletionTokens: result.CompletionTokens,
		TotalTokens:      result.TotalTokens,
		ContentChunks:    result.ContentChunks,
		Streamed:         result.Streamed,
		StreamCompleted:  result.StreamCompleted,
		ResponsePreview:  result.ResponsePreview,
		ErrorMessage:     result.ErrorMessage,
		CreatedAt:        result.CreatedAt,
	}
}

