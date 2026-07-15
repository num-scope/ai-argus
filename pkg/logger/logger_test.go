package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInitConfiguresZapLevel(t *testing.T) {
	if err := Init("debug", "console"); err != nil {
		t.Fatalf("init logger: %v", err)
	}
	t.Cleanup(Sync)
	if !L().Core().Enabled(zap.DebugLevel) {
		t.Fatal("expected debug logging to be enabled")
	}
}

func TestInitRejectsInvalidLevel(t *testing.T) {
	if err := Init("verbose", "console"); err == nil {
		t.Fatal("expected invalid log level error")
	}
}

func TestInitRejectsInvalidFormat(t *testing.T) {
	if err := Init("info", "pretty"); err == nil {
		t.Fatal("expected invalid log format error")
	}
}

func TestConsoleFormatIsReadable(t *testing.T) {
	var output bytes.Buffer
	built, err := buildLogger(zapcore.InfoLevel, "console", zapcore.AddSync(&output), zapcore.AddSync(&output))
	if err != nil {
		t.Fatalf("build logger: %v", err)
	}
	built.Info("server started", zap.String("address", "127.0.0.1:8080"))
	line := output.String()
	for _, expected := range []string{" INFO ", "server started", `"service": "ai-argus"`, `"address": "127.0.0.1:8080"`} {
		if !strings.Contains(line, expected) {
			t.Fatalf("console output does not contain %q: %s", expected, line)
		}
	}
}

func TestJSONFormatRemainsStructured(t *testing.T) {
	var output bytes.Buffer
	built, err := buildLogger(zapcore.InfoLevel, "json", zapcore.AddSync(&output), zapcore.AddSync(&output))
	if err != nil {
		t.Fatalf("build logger: %v", err)
	}
	built.Info("server started", zap.String("address", "127.0.0.1:8080"))
	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode JSON log: %v", err)
	}
	if entry["message"] != "server started" || entry["service"] != "ai-argus" || entry["address"] != "127.0.0.1:8080" {
		t.Fatalf("unexpected JSON log: %#v", entry)
	}
}
