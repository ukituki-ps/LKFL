package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lkfl/internal/app"
)

func main() {
	srv := app.New()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("[server] starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[server] failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("[server] shutting down...")

	ctx, cancel := signal.NotifyContext(stop, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[server] shutdown error: %v", err)
	}
	log.Println("[server] stopped")
}
