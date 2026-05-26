# T1701 — Инициализация go.mod

## Веха

M17-bootstrap

## Тип

code

## Контекст

Первая задача проекта — инициализация Go модуля со всеми production-зависимостями.
Список зависимостей берём из `doc/архитектура/стек.md` (строка 20 — Backend таблица).

## Что сделать

Создать `go.mod` со всеми зависимостями из `стек.md`. Версии — строго по таблице:

| Модуль | Версия | Назначение |
|--------|--------|------------|
| go-chi/chi/v5 | 5.2.2 | HTTP router |
| jackc/pgx/v5 | 5.7 | PostgreSQL driver + pool |
| go-redis/v9 | 9.7.0 | Redis client |
| coreos/go-oidc | 2.3.0 | Keycloak OIDC verification |
| spf13/viper | 1.19.0 | Config (ENV + file) |
| go-playground/validator | 13.0.0 | Request validation |
| google/cel-go | 0.20.0 | CEL Rule Engine |
| google.golang.org/grpc | 1.71.0 | gRPC mono ↔ proxy |
| google.golang.org/protobuf | 1.36.0 | Protobuf codegen |
| prometheus/client_golang | 1.20.0 | Metrics |
| sentry-go | 0.27.0 | Error tracking |
| golang.org/x/crypto | 0.28.0 | bcrypt, AES-GCM |
| excelize | 2.8.0 | XLSX parsing |
| jpillora/gofpdf | 1.10.0 | PDF generation |
| rs/cors | 1.11.1 | CORS policy |

**Не добавлять:** testcontainers, k6, vitest — они в dev tooling, не в go.mod.

## Требования

- Go 1.22 (см. `стек.md` строка 21)
- Один `go.mod` на весь проект (mono-repo)
- `go mod tidy` после добавления зависимостей
- `go.sum` — committed в репозиторий

## Критерии приёмки

- [ ] `go.mod` создан, Go version 1.22
- [ ] Все 15 зависимостей добавлены с версиями из `стек.md`
- [ ] `go mod tidy` выполняется без ошибок
- [ ] `go.sum` присутствует и закоммичен
- [ ] `go build ./...` компилируется (пустой проект, но без ошибок)
