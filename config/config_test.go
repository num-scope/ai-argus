package config

import (
	"strings"
	"testing"
)

func TestLoadUsesIndependentGormLogLevel(t *testing.T) {
	t.Setenv("ARGUS_ADDRESS", "127.0.0.1:8080")
	t.Setenv("ARGUS_DATABASE_PATH", "data/test.db")
	t.Setenv("ARGUS_LOG_LEVEL", "debug")
	t.Setenv("ARGUS_LOG_FORMAT", "json")
	t.Setenv("ARGUS_GORM_LOG_LEVEL", "silent")
	t.Setenv("ARGUS_MAX_CONCURRENCY", "20")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.LogLevel != "debug" || cfg.LogFormat != "json" || cfg.GormLogLevel != "silent" {
		t.Fatalf("logging configuration was not loaded independently: %#v", cfg)
	}
}

func TestLoadRejectsInvalidLogFormat(t *testing.T) {
	t.Setenv("ARGUS_LOG_FORMAT", "pretty")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "ARGUS_LOG_FORMAT") {
		t.Fatalf("expected log format validation, got %v", err)
	}
}

func TestLoadRejectsInvalidGormLogLevel(t *testing.T) {
	t.Setenv("ARGUS_GORM_LOG_LEVEL", "verbose")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "ARGUS_GORM_LOG_LEVEL") {
		t.Fatalf("expected GORM log level validation, got %v", err)
	}
}

func TestLoadRejectsInvalidMaxConcurrency(t *testing.T) {
	t.Setenv("ARGUS_MAX_CONCURRENCY", "many")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "ARGUS_MAX_CONCURRENCY") {
		t.Fatalf("expected max concurrency validation, got %v", err)
	}
}
