// Package logger — структурированный JSON-логгер для LKFL.
//
// Обёртка над log/slog с кастомным handler'ом, добавляющим
// сервисные атрибуты (svc) в каждый лог-запись.
//
// Формат вывода (JSON):
//
//	{"ts":"2025-01-01T12:00:00Z","level":"info","svc":"lkfl-server",
//	 "tenant_id":"sdek","user_id":"uuid","msg":"catalog query executed",
//	 "duration_ms":42,"trace_id":"uuid"}
//
// Дополнительные атрибуты (tenant_id, user_id, trace_id) добавляются
// вызывающим кодом через Logger.With() или middleware.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// SvcHandler оборачивает slog.Handler и добавляет атрибут "svc"
// в каждую запись лога. Используется для идентификации сервиса
// в агрегаторах логов (Loki, Grafana).
type SvcHandler struct {
	parent slog.Handler
	svc    string
}

// Compile-time проверка реализации slog.Handler.
var _ slog.Handler = (*SvcHandler)(nil)

// NewSvcHandler создаёт обёртку над handler'ом с фиксированным
// именем сервиса.
func NewSvcHandler(parent slog.Handler, svc string) *SvcHandler {
	return &SvcHandler{parent: parent, svc: svc}
}

// Enabled проверяет, включён ли данный уровень логирования.
func (h *SvcHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.parent.Enabled(ctx, level)
}

// Handle добавляет атрибут svc в каждую запись и делегирует
// родительскому handler'у.
func (h *SvcHandler) Handle(ctx context.Context, r slog.Record) error {
	// Добавляем svc в запись через AddAttr.
	r.AddAttrs(slog.String("svc", h.svc))
	return h.parent.Handle(ctx, r)
}

// WithAttrs добавляет атрибуты в обёртку и делегирует родительскому handler'у.
func (h *SvcHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SvcHandler{
		parent: h.parent.WithAttrs(attrs),
		svc:    h.svc,
	}
}

// WithGroup создаёт дочерний handler с группой атрибутов.
func (h *SvcHandler) WithGroup(name string) slog.Handler {
	return &SvcHandler{
		parent: h.parent.WithGroup(name),
		svc:    h.svc,
	}
}

// Options — параметры создания логгера.
type Options struct {
	// Level — уровень логирования: debug, info, warn, error.
	// По умолчанию: info.
	Level string
	// Format — формат вывода: json (по умолчанию), text.
	Format string
	// Service — имя сервиса для атрибута svc.
	Service string
	// Writer — вывод логов. По умолчанию: os.Stdout.
	Writer io.Writer
}

// New создаёт структурированный JSON-логгер.
//
// При Format == "text" используется slog.TextHandler,
// иначе slog.JSONHandler (для production).
//
// Логгер всегда оборачивается в SvcHandler для добавления
// атрибута svc в каждую запись.
func New(opts Options) *slog.Logger {
	level := parseLevel(opts.Level)
	writer := opts.Writer
	if writer == nil {
		writer = os.Stdout
	}

	var parent slog.Handler
	if strings.EqualFold(opts.Format, "text") {
		parent = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		parent = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: level,
		})
	}

	svc := opts.Service
	if svc == "" {
		svc = "lkfl"
	}

	return slog.New(NewSvcHandler(parent, svc))
}

// parseLevel преобразует строковый уровень в slog.Level.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
