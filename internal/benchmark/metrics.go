package benchmark

import (
	"sort"
	"strconv"
	"time"

	"github.com/xtj/ai-argus/internal/protocol"
)

const maxLatencySamples = 10000

type PercentileMetric struct {
	Average float64 `json:"average"`
	P50     float64 `json:"p50"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
	Samples int     `json:"samples"`
}

type Summary struct {
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

type Collector struct {
	startedAt         time.Time
	total             int
	success           int
	retries           int
	completeResponses int
	usageResponses    int
	promptTokens      int
	completionTokens  int
	totalTokens       int
	generationTokens  int
	generationMS      float64
	statusCounts      map[string]int
	e2e               sampleSet
	queue             sampleSet
	request           sampleSet
	ttft              sampleSet
	tpot              sampleSet
}

type sampleSet struct {
	values []float64
	next   int
}

func NewCollector(startedAt time.Time) *Collector {
	return &Collector{startedAt: startedAt, statusCounts: make(map[string]int)}
}

func (c *Collector) Add(result protocol.Result) {
	c.total++
	c.retries += max(result.Attempts-1, 0)
	status := "无响应"
	if result.Status != nil {
		status = strconv.Itoa(*result.Status)
		if result.StreamCompleted {
			c.completeResponses++
		}
	}
	c.statusCounts[status]++
	if !result.OK {
		return
	}

	c.success++
	c.e2e.add(result.ElapsedMS)
	c.queue.add(result.QueueMS)
	c.request.add(result.RequestMS)
	if result.TTFTMS != nil {
		c.ttft.add(*result.TTFTMS)
	}
	if result.TPOTMS != nil {
		c.tpot.add(*result.TPOTMS)
	}
	if result.TotalTokens != nil {
		c.usageResponses++
		c.promptTokens += valueOrZero(result.PromptTokens)
		c.completionTokens += valueOrZero(result.CompletionTokens)
		c.totalTokens += *result.TotalTokens
		if result.TTFTMS != nil && result.CompletionTokens != nil {
			c.generationTokens += max(*result.CompletionTokens-1, 0)
			c.generationMS += max(result.RequestMS-*result.TTFTMS, 0)
		}
	}
}

func (c *Collector) Summary(finishedAt time.Time) Summary {
	duration := finishedAt.Sub(c.startedAt).Seconds()
	if duration < 0 {
		duration = 0
	}
	summary := Summary{
		DurationSeconds:   duration,
		Retries:           c.retries,
		CompleteResponses: c.completeResponses,
		StatusCounts:      cloneCounts(c.statusCounts),
		PromptTokens:      c.promptTokens,
		CompletionTokens:  c.completionTokens,
		TotalTokens:       c.totalTokens,
		E2E:               c.e2e.metric(),
		Queue:             c.queue.metric(),
		Request:           c.request.metric(),
		TTFT:              c.ttft.metric(),
		TPOT:              c.tpot.metric(),
	}
	if duration > 0 {
		summary.RPS = float64(c.total) / duration
		summary.CompletionTokensPerSec = float64(c.completionTokens) / duration
	}
	if c.total > 0 {
		summary.SuccessRate = float64(c.success) / float64(c.total) * 100
	}
	if c.success > 0 {
		summary.UsageCoverage = float64(c.usageResponses) / float64(c.success) * 100
	}
	if c.generationMS > 0 {
		summary.GenerationTokensPerSec = float64(c.generationTokens) / (c.generationMS / 1000)
	}
	return summary
}

func (s *sampleSet) add(value float64) {
	if len(s.values) < maxLatencySamples {
		s.values = append(s.values, value)
		return
	}
	s.values[s.next] = value
	s.next = (s.next + 1) % maxLatencySamples
}

func (s sampleSet) metric() PercentileMetric {
	if len(s.values) == 0 {
		return PercentileMetric{}
	}
	values := append([]float64(nil), s.values...)
	sort.Float64s(values)
	total := 0.0
	for _, value := range values {
		total += value
	}
	return PercentileMetric{
		Average: total / float64(len(values)),
		P50:     percentile(values, 0.50),
		P95:     percentile(values, 0.95),
		P99:     percentile(values, 0.99),
		Samples: len(values),
	}
}

func percentile(values []float64, percent float64) float64 {
	position := float64(len(values)-1) * percent
	lower := int(position)
	upper := min(lower+1, len(values)-1)
	fraction := position - float64(lower)
	return values[lower] + (values[upper]-values[lower])*fraction
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func cloneCounts(source map[string]int) map[string]int {
	result := make(map[string]int, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
