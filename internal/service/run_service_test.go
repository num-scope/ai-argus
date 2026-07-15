package service_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/xtj/ai-argus/database"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/dao"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/internal/model"
	"github.com/xtj/ai-argus/internal/service"
	"github.com/xtj/ai-argus/migrations"
)

func TestRunLifecyclePersistsResultsAndSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"pong\"},\"finish_reason\":\"stop\"}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":2,\"completion_tokens\":1,\"total_tokens\":3}}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "argus.db")
	if err := database.Init(path, "error"); err != nil {
		t.Fatalf("init database: %v", err)
	}
	defer database.Close()
	if err := migrations.Run(); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	target := &model.Target{
		Name: "mock-openai", Protocol: "openai", URL: server.URL, Model: "mock-model",
		ExtraHeadersJSON: "{}", ExtraBodyJSON: "{}",
	}
	if err := dao.CreateTarget(context.Background(), target); err != nil {
		t.Fatalf("create target: %v", err)
	}
	scenario := &model.Scenario{
		Name: "smoke", PromptsJSON: `["ping"]`, Concurrency: 2, TotalRequests: 3,
		Temperature: 0.7, TopP: 1, MaxOutputTokens: 16, IncludeUsage: true,
		ConnectTimeoutSeconds: 1, SaveResponsePreview: true,
	}
	if err := dao.CreateScenario(context.Background(), scenario); err != nil {
		t.Fatalf("create scenario: %v", err)
	}

	service.ConfigureRuns(context.Background(), 10)
	run, err := service.StartRun(context.Background(), dto.StartRunRequest{TargetID: target.ID, ScenarioID: scenario.ID})
	if err != nil {
		t.Fatalf("start run: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		stored, getErr := dao.GetRunByID(context.Background(), run.ID)
		if getErr != nil {
			t.Fatalf("get run: %v", getErr)
		}
		if stored.Status == model.RunStatusCompleted {
			break
		}
		if stored.Status == model.RunStatusFailed {
			t.Fatalf("run failed: %s", stored.ErrorMessage)
		}
		time.Sleep(20 * time.Millisecond)
	}

	detail, err := service.GetRunDetail(context.Background(), run.ID, 10)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if detail.Run.Status != model.RunStatusCompleted || detail.Run.Completed != 3 || len(detail.Results) != 3 {
		t.Fatalf("unexpected run detail: status=%s completed=%d results=%d", detail.Run.Status, detail.Run.Completed, len(detail.Results))
	}
	if detail.Summary.SuccessRate != 100 || detail.Summary.TotalTokens != 9 || detail.Summary.E2E.Samples != 3 {
		t.Fatalf("unexpected summary: %#v", detail.Summary)
	}
	if err := service.CancelRun(context.Background(), run.ID); !errors.Is(err, common.ErrRunAlreadyFinished) {
		t.Fatalf("expected completed run conflict, got %v", err)
	}
	stored, err := dao.GetRunByID(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("get completed run: %v", err)
	}
	if stored.Status != model.RunStatusCompleted {
		t.Fatalf("completed status was overwritten: %s", stored.Status)
	}
	if err := database.DB.Model(&model.Run{}).Where("id = ?", run.ID).Update("summary_json", "{").Error; err != nil {
		t.Fatalf("corrupt summary: %v", err)
	}
	if _, err := service.GetRunDetail(context.Background(), run.ID, 10); err == nil {
		t.Fatal("expected corrupt summary to be reported")
	}
}
