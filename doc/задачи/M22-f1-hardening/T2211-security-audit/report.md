# T2211 — Security Audit — Отчёт

## Что сделано

### 1. Rate Limiting Middleware

**Файл:** `backend/shared/pkg/middleware/ratelimit.go`

Реализован rate limiter на основе Redis sliding window (Sorted Set):
- Score = timestamp (ms), Member = уникальный ID запроса
- При каждом запросе удаляем старые записи вне окна и считаем текущие
- Fail-open: если Redis недоступен, запрос пропускается
- Извлечение IP: X-Forwarded-For → X-Real-IP → RemoteAddr

**Тесты:** `backend/shared/pkg/middleware/ratelimit_test.go`
- TestExtractIP: 6 тестов (все прошли)
- TestRateLimiterFailOpen: fail-open поведение подтверждено
- TestIndexOf: вспомогательная функция

### 2. Rate Limiting в Server

**Файл:** `backend/internal/app/server.go`

Три rate limiter'а на разных группах маршрутов:
- **Auth endpoints** (`/api/v1/auth/`): 10 запросов/минуту на IP
- **Catalog endpoints** (`/api/v1/engagements`): 100 запросов/минуту на IP
- **Admin endpoints** (`/admin/`): 60 запросов/минуту на IP

### 3. CORS Policy

**Файл:** `backend/internal/app/server.go`, `backend/internal/app/config.go`

- Заменён кастомный CORS middleware (wildcard `*`) на `rs/cors`
- Allowed origins: настраиваемые через `CORS_ALLOWED_ORIGINS` (по умолчанию `http://localhost:5173`)
- Allowed methods: GET, POST, PATCH, DELETE, OPTIONS
- Allowed headers: Accept, Authorization, Content-Type, X-Tenant-ID
- Credentials: true
- MaxAge: настраиваемый через `CORS_MAX_AGE` (по умолчанию 3600 секунд)

### 4. Security Audit Script

**Файл:** `scripts/security-audit.sh`

Автоматический скрипт проверки OWASP Top 10:
- A01: Broken Access Control — RBAC, tenant isolation, JWT middleware
- A02: Cryptographic Failures — JWT verifier, OIDC, hardcoded secrets
- A03: Injection — SQL injection (parameterized queries check)
- A05: Security Misconfiguration — CORS, rate limiting, config
- A06: Vulnerable Components — go mod tidy, npm audit
- A07: Identification Failures — JWT, sessions, CSRF state, rate limiting
- A08: Integrity Failures — go.sum, go.mod
- A09: Security Logging — auth events, logger, Prometheus, Sentry
- A10: SSRF — HTTP client calls check
- XSS: dangerouslySetInnerHTML check, React JSX escaping
- Build: `go build ./...`

**Результат:** 32 passed, 0 failed, 3 warnings (govulncheck, .env.staging, npm audit)

### 5. SQL Injection Check

Проверены все repository файлы:
- `backend/internal/engagement/catalog/repository.go` — ✅ все запросы parameterized
- `backend/internal/user/repository.go` — ✅ все запросы parameterized
- `backend/internal/tenant/repository.go` — ✅ все запросы parameterized

Динамическое построение запросов использует `fmt.Sprintf` только для позиций плейсхолдеров (`$%d`), значения всегда передаются как параметры.

### 6. XSS Check

Проверен frontend:
- Нет использования `dangerouslySetInnerHTML`
- React автоматически экранирует JSX content
- Все user-generated content рендерится через JSX

### 7. Config Changes

**Файл:** `backend/internal/app/config.go`

Добавлены новые секции конфигурации:
- `CORSConfig` — `CORS_ALLOWED_ORIGINS`, `CORS_MAX_AGE`
- `SecurityConfig` — `RATE_LIMIT_AUTH`, `RATE_LIMIT_CATALOG`, `RATE_LIMIT_ADMIN`

## Изменённые файлы

| Файл | Действие |
|------|----------|
| `backend/shared/pkg/middleware/ratelimit.go` | Создан |
| `backend/shared/pkg/middleware/ratelimit_test.go` | Создан |
| `backend/internal/app/server.go` | Изменён (rate limiting, CORS, imports) |
| `backend/internal/app/config.go` | Изменён (CORSConfig, SecurityConfig) |
| `backend/internal/app/wire.go` | Изменён (передача full config) |
| `scripts/security-audit.sh` | Создан |
| `doc/задачи/M22-f1-hardening/T2211-security-audit/plan.yaml` | Обновлён |
| `doc/задачи/M22-f1-hardening/T2211-security-audit/report.md` | Создан |

## Критерии приёмки

- [x] Rate limiting middleware (Redis sliding window)
- [x] Rate limiting на auth (10 req/min), catalog (100 req/min), admin (60 req/min)
- [x] CORS policy verification (rs/cors, configurable)
- [x] `scripts/security-audit.sh` — 32 passed, 0 failed
- [x] SQL injection check (все запросы parameterized)
- [x] XSS check (React JSX escaping)
- [x] `go build ./...` — чистая компиляция
- [x] plan.yaml обновлён
- [x] report.md заполнён

## Замечания

1. `govulncheck` не установлен — рекомендуется установить для CI pipeline
2. `npm audit` пропущен — требует `npm install` в frontend
3. `trivy` пропущен — требует Docker image
4. Rate limiter использует fail-open стратегию — при недоступности Redis запросы проходят без ограничений (безопасность vs доступность)
