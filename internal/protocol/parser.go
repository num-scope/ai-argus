package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func readStream(response *http.Response, startedAt time.Time) (*responseData, error) {
	data := &responseData{
		status:    response.StatusCode,
		streamed:  true,
		completed: false,
	}
	var raw strings.Builder
	var content strings.Builder
	eventName := ""

	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 64*1024), maxResponseBytes)
	for scanner.Scan() {
		line := scanner.Text()
		if raw.Len() < maxResponseBytes {
			raw.WriteString(line)
			raw.WriteByte('\n')
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "event:") {
			eventName = strings.TrimSpace(strings.TrimPrefix(trimmed, "event:"))
			continue
		}
		if !strings.HasPrefix(trimmed, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
		if payload == "[DONE]" {
			data.completed = true
			break
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			eventName = ""
			continue
		}
		chunk, hasToken, completed := extractStreamEvent(event, eventName)
		if hasToken {
			data.contentChunks++
			if data.ttft == nil {
				value := time.Since(startedAt)
				data.ttft = &value
			}
		}
		content.WriteString(chunk)
		data.completed = data.completed || completed
		if prompt, completion, total := extractUsage(event); total != nil {
			data.promptTokens = prompt
			data.completionTokens = completion
			data.totalTokens = total
		}
		if message := extractMessage(event); message != "" && content.Len() == 0 {
			data.message = message
		}
		if eventName == "error" {
			data.businessError = extractBusinessError(event)
			if data.businessError == "" {
				data.businessError = marshalCompact(event)
			}
		}
		eventName = ""
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取流式响应: %w", err)
	}
	if content.Len() > 0 {
		data.message = content.String()
	}
	data.raw = raw.String()
	data.requestDuration = time.Since(startedAt)
	return data, nil
}

func parseJSONResponse(status int, raw []byte, duration time.Duration) *responseData {
	data := &responseData{
		status:          status,
		streamed:        false,
		completed:       true,
		requestDuration: duration,
		raw:             string(raw),
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		data.message = string(raw)
		return data
	}
	data.message = extractMessage(payload)
	data.businessError = extractBusinessError(payload)
	data.promptTokens, data.completionTokens, data.totalTokens = extractUsage(payload)
	return data
}

func extractStreamEvent(event map[string]any, eventName string) (string, bool, bool) {
	if eventName == "reasoning" || eventName == "delta" || eventName == "done" {
		content, _ := event["content"].(string)
		hasToken := content != ""
		chunk := ""
		if eventName == "delta" {
			chunk = content
		}
		return chunk, hasToken, eventName == "done" || event["finish_reason"] != nil
	}

	choices, _ := event["choices"].([]any)
	if len(choices) == 0 {
		return "", false, false
	}
	choice, _ := choices[0].(map[string]any)
	delta, _ := choice["delta"].(map[string]any)
	content := stringifyContent(delta["content"])
	reasoning, _ := delta["reasoning_content"].(string)
	return content, content != "" || reasoning != "", choice["finish_reason"] != nil
}

func extractMessage(payload map[string]any) string {
	choices, _ := payload["choices"].([]any)
	if len(choices) == 0 {
		return ""
	}
	choice, _ := choices[0].(map[string]any)
	message, _ := choice["message"].(map[string]any)
	return stringifyContent(message["content"])
}

func stringifyContent(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func extractUsage(payload map[string]any) (*int, *int, *int) {
	usage, _ := payload["usage"].(map[string]any)
	if usage == nil {
		return nil, nil, nil
	}
	prompt := numberAsInt(usage["prompt_tokens"])
	completion := numberAsInt(usage["completion_tokens"])
	total := numberAsInt(usage["total_tokens"])
	if total == nil && prompt != nil && completion != nil {
		value := *prompt + *completion
		total = &value
	}
	return prompt, completion, total
}

func extractBusinessError(payload map[string]any) string {
	if code, exists := payload["code"]; exists {
		codeText := fmt.Sprint(code)
		if codeText != "0" && codeText != "0.0" && codeText != "200" && codeText != "200.0" {
			message := firstString(payload, "msg", "message")
			if message == "" {
				message = "未知错误"
			}
			return fmt.Sprintf("业务 code=%s: %s", codeText, message)
		}
	}
	if value, exists := payload["error"]; exists && value != nil {
		if message, ok := value.(string); ok {
			return message
		}
		return marshalCompact(value)
	}
	return ""
}

func numberAsInt(value any) *int {
	switch number := value.(type) {
	case float64:
		converted := int(number)
		return &converted
	case int:
		converted := number
		return &converted
	case json.Number:
		converted, err := number.Int64()
		if err == nil {
			value := int(converted)
			return &value
		}
	}
	return nil
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func marshalCompact(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return compact(string(encoded), 500)
}
