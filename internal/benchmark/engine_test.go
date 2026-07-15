package benchmark

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xtj/ai-argus/internal/protocol"
)

func TestRunStopsWhenPhaseCallbackFails(t *testing.T) {
	expected := errors.New("persist phase")
	_, err := Run(context.Background(), Config{
		Protocol: protocol.Config{URL: "http://127.0.0.1", ConnectTimeoutSeconds: 1},
		Prompts:  []string{"prompt"}, Concurrency: 1, TotalRequests: 1,
	}, func(string) error {
		return expected
	}, func(protocol.Result) error {
		t.Fatal("result callback must not run")
		return nil
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected phase error, got %v", err)
	}
}

func TestRunMeasuresQueueWait(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(40 * time.Millisecond)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	queueValues := make([]float64, 0, 3)
	_, err := Run(context.Background(), Config{
		Protocol: protocol.Config{
			URL: server.URL, Protocol: "openai", Model: "model",
			TopP: 1, MaxOutputTokens: 8, ConnectTimeoutSeconds: 1,
		},
		Prompts: []string{"prompt"}, Concurrency: 1, TotalRequests: 3,
	}, func(string) error {
		return nil
	}, func(result protocol.Result) error {
		queueValues = append(queueValues, result.QueueMS)
		return nil
	})
	if err != nil {
		t.Fatalf("run benchmark: %v", err)
	}
	if len(queueValues) != 3 {
		t.Fatalf("expected three queue samples, got %d", len(queueValues))
	}
	if queueValues[1] < 20 {
		t.Fatalf("expected queued request to wait, got %.2fms", queueValues[1])
	}
}
