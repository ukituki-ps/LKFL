# T1706 — cmd/server/main.go (HTTP entry point)

## Веха

M17-bootstrap

## Тип

code

## Контекст

`cmd/server/main.go` — точка входа lkfl-server.
Инициализирует DI, запускает HTTP server, обрабатывает graceful shutdown.

Описан в `doc/архитектура/модули.md` строка 28 (`cmd/server/ — lkfl-server — HTTP API, бизнес-логика`).

## Что сделать

### `cmd/server/main.go`

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "lkfl/app"
)

func main() {
    // 1. Load config
    cfg, err := app.LoadConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, "config error: %v\n", err)
        os.Exit(1)
    }

    // 2. Provide dependencies
    server, cleanup, err := app.Provide(cfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "init error: %v\n", err)
        os.Exit(1)
    }
    defer cleanup()

    // 3. Start server in goroutine
    errCh := make(chan error, 1)
    go func() {
        errCh <- server.Start()
    }()

    // 4. Wait for interrupt signal (graceful shutdown)
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-errCh:
        fmt.Fprintf(os.Stderr, "server error: %v\n", err)
        os.Exit(1)
    case <-quit:
        fmt.Println("shutting down...")
    }

    // 5. Graceful shutdown (30s timeout)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        fmt.Fprintf(os.Stderr, "shutdown error: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("server stopped")
}
```

### `cmd/server/go.mod` directive

В корневом `go.mod` — `cmd/server` использует один `go.mod` (mono-repo).

## Требования

- Graceful shutdown — SIGINT/SIGTERM → 30s timeout → close connections
- Error handling — os.Stderr + exit code 1
- Config validation — fail fast при запуске
- No panics — все errors возвращаются, не panic-ат

## Критерии приёмки

- [ ] `cmd/server/main.go` создан
- [ ] `go run ./cmd/server/` запускает server на порту из config
- [ ] `curl http://localhost:8080/healthz` → 200 OK
- [ ] SIGINT → graceful shutdown (30s timeout)
- [ ] Ошибка config → exit 1 с сообщением в stderr
- [ ] Ошибка DB connection → exit 1 с сообщением в stderr
