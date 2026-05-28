package logger_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"lkfl/shared/pkg/logger"
)

func TestNew_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.Options{
		Level:   "info",
		Format:  "json",
		Service: "lkfl-server",
		Writer:  &buf,
	})

	l.Info("test message", "key", "value")

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected non-empty log output")
	}

	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Проверяем обязательные поля.
	if entry["svc"] != "lkfl-server" {
		t.Errorf("expected svc=lkfl-server, got %v", entry["svc"])
	}
	if entry["level"] != "INFO" {
		t.Errorf("expected level=INFO, got %v", entry["level"])
	}
	if entry["msg"] != "test message" {
		t.Errorf("expected msg='test message', got %v", entry["msg"])
	}
	if entry["key"] != "value" {
		t.Errorf("expected key=value, got %v", entry["key"])
	}
	// time поле присутствует (slog добавляет его автоматически).
	if _, ok := entry["time"]; !ok {
		t.Error("expected 'time' field in JSON output")
	}
}

func TestNew_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.Options{
		Level:   "info",
		Format:  "text",
		Service: "lkfl-server",
		Writer:  &buf,
	})

	l.Info("test message")

	line := strings.TrimSpace(buf.String())
	if !strings.Contains(line, "lkfl-server") {
		t.Errorf("expected svc in text output, got: %s", line)
	}
	if !strings.Contains(line, "test message") {
		t.Errorf("expected msg in text output, got: %s", line)
	}
}

func TestNew_DefaultService(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.Options{
		Level:  "info",
		Format: "json",
		Writer: &buf,
	})

	l.Info("test")

	line := strings.TrimSpace(buf.String())
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if entry["svc"] != "lkfl" {
		t.Errorf("expected default svc=lkfl, got %v", entry["svc"])
	}
}

func TestNew_LogLevels(t *testing.T) {
	tests := []struct {
		level    string
		call     func(l *slog.Logger)
		expected bool
	}{
		{"debug", func(l *slog.Logger) { l.Debug("msg") }, true},
		{"info", func(l *slog.Logger) { l.Info("msg") }, true},
		{"warn", func(l *slog.Logger) { l.Warn("msg") }, true},
		{"error", func(l *slog.Logger) { l.Error("msg") }, true},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			var buf bytes.Buffer
			l := logger.New(logger.Options{
				Level:   tt.level,
				Format:  "json",
				Service: "test",
				Writer:  &buf,
			})

			tt.call(l)

			if buf.Len() == 0 {
				t.Errorf("expected log output for level %s", tt.level)
			}
		})
	}
}

func TestNew_FilteredByLevel(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.Options{
		Level:   "warn",
		Format:  "json",
		Service: "test",
		Writer:  &buf,
	})

	l.Info("should not appear")
	l.Warn("should appear")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 log line, got %d", len(lines))
	}
	if !strings.Contains(buf.String(), "should appear") {
		t.Error("expected warn message in output")
	}
	if strings.Contains(buf.String(), "should not appear") {
		t.Error("info message should be filtered out at warn level")
	}
}

func TestSvcHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.Options{
		Level:   "info",
		Format:  "json",
		Service: "lkfl-server",
		Writer:  &buf,
	})

	// With() создаёт дочерний logger с дополнительными атрибутами.
	child := l.With("tenant_id", "sdek")
	child.Info("action", "user_id", "uuid-123")

	line := strings.TrimSpace(buf.String())
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if entry["tenant_id"] != "sdek" {
		t.Errorf("expected tenant_id=sdek, got %v", entry["tenant_id"])
	}
	if entry["user_id"] != "uuid-123" {
		t.Errorf("expected user_id=uuid-123, got %v", entry["user_id"])
	}
	if entry["svc"] != "lkfl-server" {
		t.Errorf("expected svc=lkfl-server, got %v", entry["svc"])
	}
}
