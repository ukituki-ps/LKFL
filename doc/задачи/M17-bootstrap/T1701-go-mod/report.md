# T1701 — Отчёт о выполнении

## Статус

выполнено

## Что сделано

1. **go.mod** — модуль `lkfl`, Go version 1.22.
2. **15 production-зависимостей** добавлены в прямой `require` блок.
3. **45 транзитивных зависимостей** в `// indirect` блоке (после `go mod tidy`).
4. **go.sum** — 469 строк, все хеши.
5. **internal/bootstrap/deps.go** — stub с blank imports для удержания зависимостей в go.mod (удалится после T1702).

## Результаты

| Команда | Результат |
|---------|-----------|
| `go mod tidy` | ✅ без ошибок |
| `go build ./...` | ✅ без ошибок |

## Структура файлов

```
backend/
├── go.mod              # модуль lkfl, Go 1.22, 15 direct + 45 indirect deps
├── go.sum              # хеши всех зависимостей
└── internal/
    └── bootstrap/
        └── deps.go     # stub — blank imports всех 15 зависимостей
```

## Зависимости (15 прямых)

| Модуль | Запрошено | Реально | Примечание |
|--------|-----------|---------|------------|
| github.com/go-chi/chi/v5 | v5.2.2 | v5.2.2 | ✅ |
| github.com/jackc/pgx/v5 | v5.7.0 | v5.7.0 | ✅ |
| github.com/go-redis/redis/v9 | v9.7.0 | github.com/redis/go-redis/v9 v9.7.0 | ⚠️ модуль переименован |
| github.com/coreos/go-oidc | v2.3.0 | v2.3.0+incompatible | ⚠️ без v2 subdirectory |
| github.com/spf13/viper | v1.19.0 | v1.19.0 | ✅ |
| github.com/go-playground/validator/v10 | v13.0.0 | v10.27.0 | ⚠️ v13 не существует для пути /v10 |
| github.com/google/cel-go | v0.20.0 | v0.20.0 | ✅ |
| google.golang.org/grpc | v1.71.0 | v1.71.0 | ✅ |
| google.golang.org/protobuf | v1.36.0 | v1.36.4 | ⚠️ grpc требует >= v1.36.4 |
| github.com/prometheus/client_golang | v1.20.0 | v1.20.0 | ✅ |
| github.com/getsentry/sentry-go | v0.27.0 | v0.27.0 | ✅ |
| golang.org/x/crypto | v0.28.0 | v0.33.0 | ⚠️ validator требует >= v0.33.0 |
| github.com/xuri/excelize/v2 | v2.8.0 | v2.8.0 | ✅ |
| github.com/jpillora/gofpdf | v1.10.0 | github.com/jung-kurt/gofpdf v1.10.0 | ⚠️ репозиторий перенесён |
| github.com/rs/cors | v1.11.1 | v1.11.1 | ✅ |

## Отклонения от brief.md

6 зависимостей с отклонениями — все обусловлены реальным состоянием экосистемы Go:
- **Redis** — модуль переименован владельцами
- **go-oidc** — не следует Go module versioning conventions
- **validator** — v13 не существует для пути /v10
- **protobuf** — grpc v1.71.0 требует >= v1.36.4
- **crypto** — validator v10.27.0 требует >= v0.33.0
- **gofpdf** — оригинальный репозиторий удалён, проект перенесён

## Следующие шаги

T1702 — создание структуры проекта (cmd/, internal/, shared/)
