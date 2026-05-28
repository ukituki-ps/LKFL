package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DeployRequest — тело запроса на деплой.
type DeployRequest struct {
	Branch   string `json:"branch"`
	SHA      string `json:"sha"`
	PR       int    `json:"pr"`
	ImageTag string `json:"imageTag"`
}

type handler struct {
	deployer *deployer
	cfg      Config
}

func newHandler(d *deployer, cfg Config) *handler {
	return &handler{deployer: d, cfg: cfg}
}

// validateAuth проверяет Bearer token в Authorization header.
// В dev-режиме (пустой WEBHOOK_SECRET) авторизация пропускается.
func (h *handler) validateAuth(r *http.Request) bool {
	if h.cfg.WebhookSecret == "" {
		return true // dev mode — без авторизации
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	return parts[1] == h.cfg.WebhookSecret
}

// writeJSON кодирует val в JSON и пишет в ResponseWriter.
func writeJSON(w http.ResponseWriter, val any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(val)
}

// writeJSONStatus пишет статус и JSON в ResponseWriter.
func writeJSONStatus(w http.ResponseWriter, status int, val any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(val)
}

// conflictResponse возвращает JSON-ответ 409 Conflict.
func conflictResponse(w http.ResponseWriter, message string) {
	_ = writeJSONStatus(w, http.StatusConflict, map[string]string{
		"status":  "conflict",
		"message": message,
	})
}

// handleDeploy обрабатывает POST /deploy — запуск деплоя.
func (h *handler) handleDeploy(w http.ResponseWriter, r *http.Request) {
	if !h.validateAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	var req DeployRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Проверяем, что нет активного деплоя
	if !h.deployer.sm.canDeploy() {
		conflictResponse(w, "deploy already in progress")
		return
	}

	// Atomically захватываем слот для деплоя (защита от race condition)
	if !h.deployer.sm.tryAcquire() {
		conflictResponse(w, "deploy started by concurrent request")
		return
	}

	if req.ImageTag == "" {
		req.ImageTag = fmt.Sprintf("%s-%s", req.Branch, req.SHA[:7])
	}

	_ = writeJSONStatus(w, http.StatusAccepted, map[string]string{
		"status":  "queued",
		"message": fmt.Sprintf("deploying %s (tag: %s)", req.Branch, req.ImageTag),
	})

	// Асинхронный запуск деплоя
	go h.deployer.Deploy(req)
}

// handleDeployPR обрабатывает POST /deploy/pr — деплой PR preview.
func (h *handler) handleDeployPR(w http.ResponseWriter, r *http.Request) {
	if !h.validateAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: PR preview на отдельный порт/композ
	_ = writeJSONStatus(w, http.StatusAccepted, map[string]string{
		"status":  "queued",
		"message": "PR preview deploy queued",
	})
}

// handleRollback обрабатывает POST /rollback — откат к предыдущему тегу.
func (h *handler) handleRollback(w http.ResponseWriter, r *http.Request) {
	if !h.validateAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.deployer.sm.canDeploy() {
		conflictResponse(w, "deploy already in progress")
		return
	}

	// Atomically захватываем слот для rollback (защита от race condition)
	if !h.deployer.sm.tryAcquire() {
		conflictResponse(w, "rollback started by concurrent request")
		return
	}

	_ = writeJSONStatus(w, http.StatusAccepted, map[string]string{
		"status":  "queued",
		"message": "rollback queued",
	})

	go h.deployer.Rollback()
}

// handleStatus обрабатывает GET /status — текущее состояние деплоя.
func (h *handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	state := h.deployer.sm.getState()
	_ = writeJSON(w, state)
}

// handleLogs обрабатывает GET /logs — логи последнего деплоя.
func (h *handler) handleLogs(w http.ResponseWriter, r *http.Request) {
	logs := h.deployer.sm.getLogs()
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprint(w, logs)
}

// handleHealthz обрабатывает GET /healthz — health check эндпоинт.
func (h *handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ok")
}
