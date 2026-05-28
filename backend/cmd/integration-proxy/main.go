// Package main — lkfl-integration-proxy stub.
//
// В F1 — заглушка gRPC сервера для docker compose и локальной разработки.
// В F3 (M30+) — полноценный proxy с 11 провайдерами льгот.
//
// Порты:
//
//	:8090 — gRPC (mono ↔ proxy)
//	:8091 — HTTP webhooks (от провайдеров)
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	port := os.Getenv("PROXY_GRPC_PORT")
	if port == "" {
		port = "8090"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen :%s: %v", port, err)
	}

	s := grpc.NewServer()

	// gRPC health check service (стандартный)
	// Поддерживает grpc_health_probe и docker-compose healthcheck.
	h := health.NewServer()
	h.SetServingStatus(
		"lkfl.integration.proxy",
		grpc_health_v1.HealthCheckResponse_SERVING,
	)
	grpc_health_v1.RegisterHealthServer(s, h)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Println("integration-proxy stub listening on :" + port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	<-quit
	fmt.Println("shutting down proxy...")

	// GracefulStop без контекста (gRPC v1.71).
	// Для таймаута используем goroutine + Stop.
	done := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("proxy stopped")
	case <-time.After(10 * time.Second):
		fmt.Println("proxy force stopped (timeout)")
		s.Stop()
	}
}
