package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/dao"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/internal/model"
	"github.com/xtj/ai-argus/internal/utils"
)

func CreateTarget(ctx context.Context, req dto.TargetRequest) (*dto.TargetResponse, error) {
	normalizeTargetRequest(&req)
	if err := validateTargetRequest(req); err != nil {
		return nil, err
	}
	target := &model.Target{
		Name:               req.Name,
		Protocol:           req.Protocol,
		URL:                req.URL,
		Model:              req.Model,
		APIKey:             req.APIKey,
		CustomContentField: req.CustomContentField,
		ExtraHeadersJSON:   defaultJSONObject(req.ExtraHeadersJSON),
		ExtraBodyJSON:      defaultJSONObject(req.ExtraBodyJSON),
	}
	if err := dao.CreateTarget(ctx, target); err != nil {
		if errors.Is(err, common.ErrAlreadyExists) {
			return nil, common.ErrAlreadyExists
		}
		return nil, err
	}
	response := toTargetResponse(*target)
	return &response, nil
}

func ListTargets(ctx context.Context) ([]dto.TargetResponse, error) {
	targets, err := dao.ListTargets(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.TargetResponse, 0, len(targets))
	for _, target := range targets {
		responses = append(responses, toTargetResponse(target))
	}
	return responses, nil
}

func ListTargetsPage(ctx context.Context, req dto.PageQuery) (*dto.PageResponse[dto.TargetResponse], error) {
	page, pageSize := utils.NormalizePage(req.Page, req.PageSize)
	items, total, err := dao.ListTargetsPage(ctx, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.TargetResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toTargetResponse(item))
	}
	result := dto.NewPageResponse(responses, total, page, pageSize)
	return &result, nil
}

func DeleteTarget(ctx context.Context, id int64) error {
	return dao.DeleteTarget(ctx, id)
}

func normalizeTargetRequest(req *dto.TargetRequest) {
	req.Name = strings.TrimSpace(req.Name)
	req.Protocol = strings.ToLower(strings.TrimSpace(req.Protocol))
	req.URL = strings.TrimSpace(req.URL)
	req.Model = strings.TrimSpace(req.Model)
	req.APIKey = strings.TrimSpace(req.APIKey)
	req.CustomContentField = strings.TrimSpace(req.CustomContentField)
	req.ExtraHeadersJSON = strings.TrimSpace(req.ExtraHeadersJSON)
	req.ExtraBodyJSON = strings.TrimSpace(req.ExtraBodyJSON)
	if req.Protocol == "custom" && req.CustomContentField == "" {
		req.CustomContentField = "content"
	}
}

func validateTargetRequest(req dto.TargetRequest) error {
	if req.Name == "" || req.URL == "" || req.Model == "" {
		return invalidRequest("名称、接口地址和模型不能为空")
	}
	if req.Protocol != "openai" && req.Protocol != "custom" {
		return invalidRequest("协议仅支持 openai 或 custom")
	}
	parsed, err := url.ParseRequestURI(req.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return invalidRequest("接口地址必须是有效的 HTTP 或 HTTPS URL")
	}
	if req.Protocol == "custom" && req.CustomContentField == "" {
		return invalidRequest("自定义协议需要内容字段")
	}
	if req.Protocol == "custom" && isReservedContentField(req.CustomContentField) {
		return invalidRequest("自定义内容字段不能覆盖协议保留字段")
	}
	if err := validateStringMap(req.ExtraHeadersJSON, "额外请求头"); err != nil {
		return err
	}
	if err := validateAnyMap(req.ExtraBodyJSON, "额外请求体"); err != nil {
		return err
	}
	return nil
}

func isReservedContentField(field string) bool {
	switch field {
	case "model", "messages", "temperature", "top_p", "max_tokens", "seed", "stream", "stream_options":
		return true
	default:
		return false
	}
}

func validateStringMap(value, label string) error {
	if value == "" {
		return nil
	}
	var parsed map[string]string
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return invalidRequest(label + "必须是字符串键值 JSON 对象")
	}
	return nil
}

func validateAnyMap(value, label string) error {
	if value == "" {
		return nil
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return invalidRequest(label + "必须是 JSON 对象")
	}
	return nil
}

func defaultJSONObject(value string) string {
	if value == "" {
		return "{}"
	}
	return value
}

func toTargetResponse(target model.Target) dto.TargetResponse {
	return dto.TargetResponse{
		ID:                 target.ID,
		Name:               target.Name,
		Protocol:           target.Protocol,
		URL:                target.URL,
		Model:              target.Model,
		HasAPIKey:          target.APIKey != "",
		CustomContentField: target.CustomContentField,
		CreatedAt:          target.CreatedAt,
		UpdatedAt:          target.UpdatedAt,
	}
}
