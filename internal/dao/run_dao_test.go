package dao

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/xtj/ai-argus/database"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/model"
	"github.com/xtj/ai-argus/migrations"
)

func TestCancelRunIfActivePreservesCompletedRun(t *testing.T) {
	initTestDatabase(t)
	run := &model.Run{
		TargetID: 1, ScenarioID: 1, TargetName: "target", ScenarioName: "scenario",
		Protocol: "openai", Model: "model", Status: model.RunStatusCompleted, ConfigJSON: "{}", SummaryJSON: "{}",
	}
	if err := CreateRun(context.Background(), run); err != nil {
		t.Fatalf("create run: %v", err)
	}
	cancelled, err := CancelRunIfActive(context.Background(), run.ID, time.Now(), "cancelled")
	if err != nil {
		t.Fatalf("cancel run: %v", err)
	}
	if cancelled {
		t.Fatal("completed run must not be cancelled")
	}
	stored, err := GetRunByID(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if stored.Status != model.RunStatusCompleted {
		t.Fatalf("status changed to %s", stored.Status)
	}
}

func TestAppendRunResultsRollsBackProgressOnInsertFailure(t *testing.T) {
	initTestDatabase(t)
	run := &model.Run{
		TargetID: 1, ScenarioID: 1, TargetName: "target", ScenarioName: "scenario",
		Protocol: "openai", Model: "model", Status: model.RunStatusQueued, ConfigJSON: "{}", SummaryJSON: "{}",
	}
	if err := CreateRun(context.Background(), run); err != nil {
		t.Fatalf("create run: %v", err)
	}
	existing := model.RequestResult{ID: 1, RunID: run.ID, RequestIndex: 1, Prompt: "existing"}
	if err := database.DB.Create(&existing).Error; err != nil {
		t.Fatalf("create existing result: %v", err)
	}

	run.Status = model.RunStatusRunning
	run.Completed = 2
	results := []model.RequestResult{
		{ID: 2, RunID: run.ID, RequestIndex: 2, Prompt: "new"},
		{ID: 1, RunID: run.ID, RequestIndex: 3, Prompt: "duplicate"},
	}
	if err := AppendRunResults(context.Background(), run, results); err == nil {
		t.Fatal("expected duplicate result insert to fail")
	}
	stored, err := GetRunByID(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if stored.Status != model.RunStatusQueued || stored.Completed != 0 {
		t.Fatalf("run progress was not rolled back: %#v", stored)
	}
	var count int64
	if err := database.DB.Model(&model.RequestResult{}).Where("run_id = ?", run.ID).Count(&count).Error; err != nil {
		t.Fatalf("count results: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one existing result, got %d", count)
	}
}

func TestCreateTargetNormalizesDuplicateError(t *testing.T) {
	initTestDatabase(t)
	first := &model.Target{Name: "duplicate", Protocol: "openai", URL: "http://example.com", Model: "model"}
	if err := CreateTarget(context.Background(), first); err != nil {
		t.Fatalf("create first target: %v", err)
	}
	second := &model.Target{Name: "duplicate", Protocol: "openai", URL: "http://example.com", Model: "model"}
	if err := CreateTarget(context.Background(), second); !errors.Is(err, common.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func initTestDatabase(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "argus.db")
	if err := database.Init(path, "silent"); err != nil {
		t.Fatalf("init database: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	if err := migrations.Run(); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
}
