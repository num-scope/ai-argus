package benchmark

import (
	"math"
	"testing"
	"time"

	"github.com/xtj/ai-argus/internal/protocol"
)

func TestCollectorBuildsLatencyAndTokenSummary(t *testing.T) {
	startedAt := time.Now().Add(-2 * time.Second)
	collector := NewCollector(startedAt)
	for index, latency := range []float64{100, 200, 300, 400, 500} {
		status, promptTokens, completionTokens, totalTokens := 200, 10, 5, 15
		ttft, tpot := latency/2, 10.0
		collector.Add(protocol.Result{
			Index: index + 1, OK: true, Status: &status, Attempts: 1,
			ElapsedMS: latency, RequestMS: latency, TTFTMS: &ttft, TPOTMS: &tpot,
			PromptTokens: &promptTokens, CompletionTokens: &completionTokens, TotalTokens: &totalTokens,
			StreamCompleted: true,
		})
	}
	summary := collector.Summary(startedAt.Add(2 * time.Second))
	if summary.E2E.P50 != 300 || summary.E2E.P95 != 480 || summary.E2E.Samples != 5 {
		t.Fatalf("unexpected E2E summary: %#v", summary.E2E)
	}
	if summary.SuccessRate != 100 || summary.TotalTokens != 75 {
		t.Fatalf("unexpected aggregate summary: %#v", summary)
	}
	if math.Abs(summary.RPS-2.5) > 0.001 {
		t.Fatalf("unexpected RPS: %f", summary.RPS)
	}
}
