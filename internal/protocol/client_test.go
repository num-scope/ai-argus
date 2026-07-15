package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientExecutesOpenAIStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["model"] != "argus-test" || payload["stream"] != true {
			t.Fatalf("unexpected payload: %#v", payload)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"你\"},\"finish_reason\":null}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"好\"},\"finish_reason\":\"stop\"}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":2,\"total_tokens\":5}}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := NewClient(Config{
		URL: server.URL, Protocol: "openai", Model: "argus-test",
		Temperature: 0.7, TopP: 1, MaxOutputTokens: 32,
		IncludeUsage: true, ConnectTimeoutSeconds: 1, SaveResponsePreview: true,
	}, 1)
	defer client.Close()
	result := client.Execute(context.Background(), 1, "问候", 0)
	if !result.OK || result.ResponsePreview != "你好" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.TTFTMS == nil || result.TPOTMS == nil {
		t.Fatalf("expected streaming latency metrics: %#v", result)
	}
	if result.TotalTokens == nil || *result.TotalTokens != 5 {
		t.Fatalf("expected usage tokens: %#v", result.TotalTokens)
	}
}

func TestClientCapsRetries(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "retry", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(Config{
		URL: server.URL, Protocol: "openai", Model: "model", MaxRetries: 100,
		TopP: 1, MaxOutputTokens: 8, ConnectTimeoutSeconds: 1,
	}, 1)
	defer client.Close()
	result := client.Execute(context.Background(), 1, "prompt", 0)
	if result.OK {
		t.Fatal("retrying response must fail")
	}
	if result.Attempts != MaxRetries+1 || requests.Load() != int64(MaxRetries+1) {
		t.Fatalf("retry cap failed: attempts=%d requests=%d", result.Attempts, requests.Load())
	}
}

func TestClientExecutesCustomNamedEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["query"] != "自定义问题" {
			t.Fatalf("custom content field missing: %#v", payload)
		}
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		fmt.Fprint(w, "event: reasoning\ndata: {\"content\":\"思考\"}\n\n")
		fmt.Fprint(w, "event: delta\ndata: {\"content\":\"自定义回答\"}\n\n")
		fmt.Fprint(w, "event: done\ndata: {}\n\n")
	}))
	defer server.Close()

	client := NewClient(Config{
		URL: server.URL, Protocol: "custom", Model: "custom-model", CustomContentField: "query",
		Temperature: 0.7, TopP: 1, MaxOutputTokens: 32, ConnectTimeoutSeconds: 1, SaveResponsePreview: true,
	}, 1)
	defer client.Close()
	result := client.Execute(context.Background(), 1, "自定义问题", 0)
	if !result.OK || !result.StreamCompleted || result.ResponsePreview != "自定义回答" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.ContentChunks != 2 {
		t.Fatalf("expected reasoning and delta chunks, got %d", result.ContentChunks)
	}
}
