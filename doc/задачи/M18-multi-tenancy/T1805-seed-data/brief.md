# T1805 — Seed Data (tenant sdek)

## Веха

M18-multi-tenancy

## Тип

code

## Контекст

Seed данные для разработки — tenant «sdek» с brand config.
Не production seed — только для dev/staging окружения.

## Что сделать

### `cmd/seed/main.go`

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "os"

    "github.com/jackc/pgx/v5"
    "github.com/google/uuid"
)

func main() {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        fmt.Println("DB_DSN required")
        os.Exit(1)
    }

    conn, err := pgx.Connect(context.Background(), dsn)
    if err != nil {
        fmt.Fprintf(os.Stderr, "connect error: %v\n", err)
        os.Exit(1)
    }
    defer conn.Close(context.Background())

    // Tenant SDEK
    sdekID := uuid.New()
    _, err = conn.Exec(context.Background(), `
        INSERT INTO lkfl_platform.tenants (id, slug, name, status, settings)
        VALUES ($1, 'sdek', 'СДЭК', 'active', '{}')
    `, sdekID)
    if err != nil {
        fmt.Fprintf(os.Stderr, "insert tenant error: %v\n", err)
        os.Exit(1)
    }

    // Brand config SDEK
    _, err = conn.Exec(context.Background(), `
        INSERT INTO lkfl_platform.tenant_brand_config (
            tenant_id, primary_color, secondary_color, brand_name, css_variables
        ) VALUES ($1, '#E30613', '#FFFFFF', 'СДЭК Льготы', '{
            "--april-color-primary": "#E30613",
            "--april-color-primary-hover": "#C50510"
        }')
    `, sdekID)
    if err != nil {
        fmt.Fprintf(os.Stderr, "insert brand config error: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Seeded tenant: sdek (id=%s)\n", sdekID)
}
```

### Запуск

```bash
make seed
# или
docker compose exec lkfl-server go run ./cmd/seed/
```

## Требования

- ID генерируется при каждом запуске (не hardcoded)
- Slug: `sdek`
- Primary color: `#E30613` (СДЭК красный)
- Brand name: `СДЭК Льготы`
- CSS variables для April tokens override
- Idempotent — повторный запуск не ломает (INSERT ... ON CONFLICT)
- Только для dev/staging (не в production pipeline)

## Критерии приёмки

- [ ] `cmd/seed/main.go` создан
- [ ] `make seed` загружает tenant sdek + brand config
- [ ] Slug: `sdek`, name: `СДЭК`
- [ ] Brand config: primary_color `#E30613`, css_variables для April tokens
- [ ] Idempotent (повторный запуск не ломает)
- [ ] Tenant доступен через API после seed
