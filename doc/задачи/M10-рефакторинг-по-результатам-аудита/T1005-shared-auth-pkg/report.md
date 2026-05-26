# T1005 — Shared auth pkg — отчёт

## Статус

✅ завершена

## Что сделано

1. **`doc/архитектура/пакеты-platform.md`** — секция `internal/auth/` обновлена:
   - Описание изменено на "thin tenant wrapper over shared/pkg/auth"
   - Добавлена секция `shared/pkg/auth (M10 T1005)` с описанием файловой структуры (verifier.go, middleware.go, rbac.go)
   - Зависимости обновлены: `shared/pkg/auth` в список deps
2. **`doc/архитектура/модули.md`** — Payment Gateway auth модуль: зависимости изменены с `go-oidc` → `shared/pkg/auth`, Redis
3. **`doc/архитектура/adr/011-monorepo.md`** — структура репозитория обновлена: `shared/pkg/auth/` добавлен.

## Что НЕ трогать

- Файлы Go-кода platform/internal/auth/*.go — фактическое выделение shared пакета при реализации
- Ключи API, credentials, OIDC конфиг — остаются без изменений

## Проблемы

Нет — задача чистая документация.
