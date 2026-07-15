package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/dao"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/internal/model"
	"github.com/xtj/ai-argus/internal/protocol"
	"github.com/xtj/ai-argus/internal/utils"
)

func CreateScenario(ctx context.Context, req dto.ScenarioRequest) (*dto.ScenarioResponse, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.SystemPrompt = strings.TrimSpace(req.SystemPrompt)
	prompts := splitPrompts(req.Prompts)
	seed, err := parseSeed(req.Seed)
	if err != nil {
		return nil, err
	}
	if err := validateScenarioRequest(req, prompts, RunConcurrencyLimit()); err != nil {
		return nil, err
	}
	promptsJSON, err := json.Marshal(prompts)
	if err != nil {
		return nil, err
	}
	scenario := &model.Scenario{
		Name:                  req.Name,
		SystemPrompt:          req.SystemPrompt,
		PromptsJSON:           string(promptsJSON),
		Concurrency:           req.Concurrency,
		TotalRequests:         req.TotalRequests,
		WarmupRequests:        req.WarmupRequests,
		RampUpSeconds:         req.RampUpSeconds,
		Temperature:           req.Temperature,
		TopP:                  req.TopP,
		MaxOutputTokens:       req.MaxOutputTokens,
		Seed:                  seed,
		IncludeUsage:          req.IncludeUsage,
		TimeoutSeconds:        req.TimeoutSeconds,
		ConnectTimeoutSeconds: req.ConnectTimeoutSeconds,
		MaxRetries:            req.MaxRetries,
		RetryBaseDelaySeconds: req.RetryBaseDelaySeconds,
		SaveResponsePreview:   req.SaveResponsePreview,
	}
	if err := dao.CreateScenario(ctx, scenario); err != nil {
		if errors.Is(err, common.ErrAlreadyExists) {
			return nil, common.ErrAlreadyExists
		}
		return nil, err
	}
	response := toScenarioResponse(*scenario, prompts)
	return &response, nil
}

func ListScenarios(ctx context.Context) ([]dto.ScenarioResponse, error) {
	scenarios, err := dao.ListScenarios(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.ScenarioResponse, 0, len(scenarios))
	for _, scenario := range scenarios {
		prompts, err := decodePrompts(scenario.PromptsJSON)
		if err != nil {
			return nil, err
		}
		responses = append(responses, toScenarioResponse(scenario, prompts))
	}
	return responses, nil
}

func ListScenariosPage(ctx context.Context, req dto.PageQuery) (*dto.PageResponse[dto.ScenarioResponse], error) {
	page, pageSize := utils.NormalizePage(req.Page, req.PageSize)
	items, total, err := dao.ListScenariosPage(ctx, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.ScenarioResponse, 0, len(items))
	for _, item := range items {
		prompts, err := decodePrompts(item.PromptsJSON)
		if err != nil {
			return nil, err
		}
		responses = append(responses, toScenarioResponse(item, prompts))
	}
	result := dto.NewPageResponse(responses, total, page, pageSize)
	return &result, nil
}

func DeleteScenario(ctx context.Context, id int64) error {
	return dao.DeleteScenario(ctx, id)
}

func splitPrompts(value string) []string {
	lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if prompt := strings.TrimSpace(line); prompt != "" {
			result = append(result, prompt)
		}
	}
	return result
}

func decodePrompts(value string) ([]string, error) {
	var prompts []string
	if err := json.Unmarshal([]byte(value), &prompts); err != nil {
		return nil, err
	}
	return prompts, nil
}

func parseSeed(value string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	seed, err := strconv.Atoi(value)
	if err != nil {
		return nil, invalidRequest("Seed 必须是整数或留空")
	}
	return &seed, nil
}

func validateScenarioRequest(req dto.ScenarioRequest, prompts []string, maxConcurrency int) error {
	if req.Name == "" || len(prompts) == 0 {
		return invalidRequest("名称和至少一个提示词不能为空")
	}
	if req.Concurrency < 1 || req.Concurrency > maxConcurrency {
		return invalidRequest("并发数必须在 1 到平台上限之间")
	}
	if req.TotalRequests < 0 || req.WarmupRequests < 0 || req.RampUpSeconds < 0 {
		return invalidRequest("请求数、预热数和升压时间不能小于 0")
	}
	if !isFinite(req.RampUpSeconds) || req.RampUpSeconds > 86400 {
		return invalidRequest("升压时间必须是 0 到 86400 秒的有限数值")
	}
	if req.Temperature < 0 || req.Temperature > 2 || req.TopP <= 0 || req.TopP > 1 {
		return invalidRequest("Temperature 必须在 0 到 2，Top P 必须大于 0 且不超过 1")
	}
	if !isFinite(req.Temperature) || !isFinite(req.TopP) {
		return invalidRequest("Temperature 和 Top P 必须是有限数值")
	}
	if req.MaxOutputTokens < 1 {
		return invalidRequest("最大输出 Token 必须大于 0")
	}
	if !isFinite(req.TimeoutSeconds) || !isFinite(req.ConnectTimeoutSeconds) || req.TimeoutSeconds < 0 || req.TimeoutSeconds > 86400 || req.ConnectTimeoutSeconds <= 0 || req.ConnectTimeoutSeconds > 300 {
		return invalidRequest("总超时必须在 0 到 86400 秒，连接超时必须大于 0 且不超过 300 秒")
	}
	if !isFinite(req.RetryBaseDelaySeconds) || req.MaxRetries < 0 || req.MaxRetries > protocol.MaxRetries || req.RetryBaseDelaySeconds < 0 || req.RetryBaseDelaySeconds > 60 {
		return invalidRequest("重试次数必须在 0 到 10，基础退避必须在 0 到 60 秒")
	}
	return nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func toScenarioResponse(scenario model.Scenario, prompts []string) dto.ScenarioResponse {
	return dto.ScenarioResponse{
		ID:                    scenario.ID,
		Name:                  scenario.Name,
		SystemPrompt:          scenario.SystemPrompt,
		Prompts:               prompts,
		Concurrency:           scenario.Concurrency,
		TotalRequests:         scenario.TotalRequests,
		WarmupRequests:        scenario.WarmupRequests,
		RampUpSeconds:         scenario.RampUpSeconds,
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
		CreatedAt:             scenario.CreatedAt,
		UpdatedAt:             scenario.UpdatedAt,
	}
}
