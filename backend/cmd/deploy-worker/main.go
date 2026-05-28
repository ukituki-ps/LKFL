// Deploy Worker — лёгкий Go-сервис для оркестрации деплоя на staging сервере.
//
// Принимает webhook от GitHub Actions, пуллит Docker-образы из GHCR,
// запускает миграции, поднимает сервисы через docker compose,
// выполняет seed и healthcheck.
//
// Запуск:
//
//	GOOS=linux go build -o deploy-worker ./cmd/deploy-worker/
//	./deploy-worker
//
// Конфигурация (env vars):
//   PORT            — порт HTTP-сервера (default: 9091)
//   WEBHOOK_SECRET  — Bearer token для авторизации webhook
//   GHCR_TOKEN      — токен для авторизации в ghcr.io
//   COMPOSE_FILE    — путь к docker-compose файлу (default: docker-compose.staging.yml)
//   COMPOSE_DIR     — рабочая директория для docker compose (default: .)
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := loadConfig()

	sm := newStateManager()
	deployer := newDeployer(cfg, sm)
	handler := newHandler(deployer, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /deploy", handler.handleDeploy)
	mux.HandleFunc("POST /deploy/pr", handler.handleDeployPR)
	mux.HandleFunc("POST /rollback", handler.handleRollback)
	mux.HandleFunc("GET /status", handler.handleStatus)
	mux.HandleFunc("GET /logs", handler.handleLogs)
	mux.HandleFunc("GET /healthz", handler.handleHealthz)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("deploy-worker listening on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down...")
		_ = server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}

	log.Println("deploy-worker stopped")
}
