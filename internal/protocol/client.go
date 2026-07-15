package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const maxResponseBytes = 10 << 20
const MaxRetries = 10

const maxRetryDelay = 30 * time.Second

type Config struct {
	URL                   string            `json:"url"`
	APIKey                string            `json:"-"`
	Protocol              string            `json:"protocol"`
	Model                 string            `json:"model"`
	CustomContentField    string            `json:"custom_content_field,omitempty"`
	ExtraHeaders          map[string]string `json:"extra_headers,omitempty"`
	ExtraBody             map[string]any    `json:"extra_body,omitempty"`
	SystemPrompt          string            `json:"system_prompt,omitempty"`
	Temperature           float64           `json:"temperature"`
	TopP                  float64           `json:"top_p"`
	MaxOutputTokens       int               `json:"max_output_tokens"`
	Seed                  *int              `json:"seed,omitempty"`
	IncludeUsage          bool              `json:"include_usage"`
	TimeoutSeconds        float64           `json:"timeout_seconds"`
	ConnectTimeoutSeconds float64           `json:"connect_timeout_seconds"`
	MaxRetries            int               `json:"max_retries"`
	RetryBaseDelaySeconds float64           `json:"retry_base_delay_seconds"`
	SaveResponsePreview   bool              `json:"save_response_preview"`
}

type Result struct {
	Index            int
	Prompt           string
	OK               bool
	Status           *int
	Attempts         int
	ElapsedMS        float64
	QueueMS          float64
	RequestMS        float64
	TTFTMS           *float64
	TPOTMS           *float64
	PromptTokens     *int
	CompletionTokens *int
	TotalTokens      *int
	ContentChunks    int
	Streamed         bool
	StreamCompleted  bool
	ResponsePreview  string
	Error            string
}

type responseData struct {
	status           int
	message          string
	streamed         bool
	completed        bool
	businessError    string
	requestDuration  time.Duration
	ttft             *time.Duration
	promptTokens     *int
	completionTokens *int
	totalTokens      *int
	contentChunks    int
	raw              string
}

type Client struct {
	config Config
	http   *http.Client
}

func NewClient(cfg Config, concurrency int) *Client {
	dialer := &net.Dialer{Timeout: durationSeconds(cfg.ConnectTimeoutSeconds), KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          max(concurrency*2, 100),
		MaxIdleConnsPerHost:   max(concurrency, 10),
		MaxConnsPerHost:       max(concurrency, 1),
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	return &Client{config: cfg, http: &http.Client{Transport: transport}}
}

func (c *Client) Close() {
	if transport, ok := c.http.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}

func (c *Client) Execute(ctx context.Context, index int, prompt string, queue time.Duration) Result {
	startedAt := time.Now()
	var lastResponse *responseData
	var lastErr error
	attempts := 0

	retryLimit := min(max(c.config.MaxRetries, 0), MaxRetries)
	for attempt := 0; attempt <= retryLimit; attempt++ {
		attempts = attempt + 1
		response, err := c.post(ctx, prompt)
		if response != nil {
			lastResponse = response
		}
		if err == nil && response != nil && response.status >= 200 && response.status < 300 {
			if response.businessError != "" {
				lastErr = errors.New(response.businessError)
				break
			}
			if response.streamed && !response.completed {
				lastErr = errors.New("流式响应未收到 [DONE] 或 finish_reason 完成标记")
			} else {
				return buildResult(index, prompt, attempts, time.Since(startedAt), queue, response, nil, c.config.SaveResponsePreview)
			}
		} else if err != nil {
			lastErr = err
		} else if response != nil {
			lastErr = fmt.Errorf("HTTP %d: %s", response.status, compact(response.raw, 320))
			if !isRetryableStatus(response.status) {
				break
			}
		}

		if attempt < retryLimit {
			delay := retryDelay(c.config.RetryBaseDelaySeconds, attempt)
			if err := sleepContext(ctx, delay); err != nil {
				lastErr = err
				break
			}
		}
	}

	return buildResult(index, prompt, attempts, time.Since(startedAt), queue, lastResponse, lastErr, c.config.SaveResponsePreview)
}

func retryDelay(baseSeconds float64, attempt int) time.Duration {
	delay := durationSeconds(baseSeconds)
	for index := 0; index < attempt; index++ {
		if delay >= maxRetryDelay/2 {
			return maxRetryDelay
		}
		delay *= 2
	}
	return min(delay, maxRetryDelay)
}

func (c *Client) post(parent context.Context, prompt string) (*responseData, error) {
	payload := c.buildPayload(prompt)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	ctx := parent
	cancel := func() {}
	if c.config.TimeoutSeconds > 0 {
		ctx, cancel = context.WithTimeout(parent, durationSeconds(c.config.TimeoutSeconds))
	}
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.URL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "text/event-stream")
	if strings.TrimSpace(c.config.APIKey) != "" {
		request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.config.APIKey))
	}
	for key, value := range c.config.ExtraHeaders {
		request.Header.Set(key, value)
	}

	startedAt := time.Now()
	response, err := c.http.Do(request)
	if err != nil {
		return nil, normalizeRequestError(err)
	}
	defer response.Body.Close()

	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	if strings.Contains(contentType, "text/event-stream") {
		return readStream(response, startedAt)
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return parseJSONResponse(response.StatusCode, raw, time.Since(startedAt)), nil
}

func (c *Client) buildPayload(prompt string) map[string]any {
	messages := make([]map[string]string, 0, 2)
	if strings.TrimSpace(c.config.SystemPrompt) != "" {
		messages = append(messages, map[string]string{"role": "system", "content": c.config.SystemPrompt})
	}
	messages = append(messages, map[string]string{"role": "user", "content": prompt})
	payload := map[string]any{
		"model":       c.config.Model,
		"messages":    messages,
		"temperature": c.config.Temperature,
		"top_p":       c.config.TopP,
		"max_tokens":  c.config.MaxOutputTokens,
	}
	if c.config.Seed != nil {
		payload["seed"] = *c.config.Seed
	}
	for key, value := range c.config.ExtraBody {
		payload[key] = value
	}
	payload["stream"] = true
	if c.config.Protocol == "custom" {
		field := strings.TrimSpace(c.config.CustomContentField)
		if field == "" {
			field = "content"
		}
		payload[field] = prompt
	}
	if c.config.IncludeUsage {
		streamOptions, _ := payload["stream_options"].(map[string]any)
		if streamOptions == nil {
			streamOptions = make(map[string]any)
		}
		streamOptions["include_usage"] = true
		payload["stream_options"] = streamOptions
	}
	return payload
}

func buildResult(index int, prompt string, attempts int, elapsed, queue time.Duration, response *responseData, err error, savePreview bool) Result {
	result := Result{Index: index, Prompt: prompt, Attempts: attempts, ElapsedMS: milliseconds(elapsed), QueueMS: milliseconds(queue)}
	if response != nil {
		status := response.status
		result.Status = &status
		result.RequestMS = milliseconds(response.requestDuration)
		result.PromptTokens = response.promptTokens
		result.CompletionTokens = response.completionTokens
		result.TotalTokens = response.totalTokens
		result.ContentChunks = response.contentChunks
		result.Streamed = response.streamed
		result.StreamCompleted = response.completed
		if response.ttft != nil {
			value := milliseconds(*response.ttft)
			result.TTFTMS = &value
		}
		if result.TTFTMS != nil && result.CompletionTokens != nil && *result.CompletionTokens > 1 {
			value := max(result.RequestMS-*result.TTFTMS, 0) / float64(*result.CompletionTokens-1)
			result.TPOTMS = &value
		}
		if savePreview {
			result.ResponsePreview = compact(response.message, 500)
		}
	}
	if err != nil {
		result.Error = compact(err.Error(), 500)
		return result
	}
	result.OK = true
	return result
}

func durationSeconds(value float64) time.Duration {
	return time.Duration(value * float64(time.Second))
}

func milliseconds(value time.Duration) float64 {
	return float64(value.Microseconds()) / 1000
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusConflict, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func normalizeRequestError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.New("请求超时")
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	return fmt.Errorf("请求失败: %w", err)
}

func compact(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len([]rune(value)) <= limit {
		return value
	}
	return string([]rune(value)[:limit]) + "..."
}
