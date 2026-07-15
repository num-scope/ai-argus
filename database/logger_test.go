package database

import (
	"context"
	"testing"

	gormLogger "gorm.io/gorm/logger"
)

func TestNewGormLoggerUsesConfiguredLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  gormLogger.LogLevel
	}{
		{name: "silent", input: "silent", want: gormLogger.Silent},
		{name: "error", input: "error", want: gormLogger.Error},
		{name: "warn", input: "warn", want: gormLogger.Warn},
		{name: "info", input: "info", want: gormLogger.Info},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newGormLogger(tt.input).(*zapGormLogger)
			if got.level != tt.want {
				t.Fatalf("expected level %d, got %d", tt.want, got.level)
			}
		})
	}
}

func TestGormLoggerRemovesQueryParameters(t *testing.T) {
	adapter := newGormLogger("info").(*zapGormLogger)
	sql, params := adapter.ParamsFilter(context.Background(), "INSERT INTO targets (api_key) VALUES (?)", "secret")
	if sql != "INSERT INTO targets (api_key) VALUES (?)" {
		t.Fatalf("unexpected SQL: %s", sql)
	}
	if len(params) != 0 {
		t.Fatalf("expected query parameters to be removed, got %#v", params)
	}
}

var _ gormLogger.Interface = (*zapGormLogger)(nil)
