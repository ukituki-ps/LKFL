# T1004 — Split api/ → public router + admin router — отчёт

## Статус

✅ завершена

## Что сделано

1. **`doc/архитектура/пакеты-platform.md`** — секция `internal/api/` полностью переписана:
   - Один router + один HandlerDeps (12 полей) → 2 router'а + 2 HandlerDeps
   - `PublicHandlerDeps` (10 полей) + `AdminHandlerDeps` (7 полей)
   - 2 middleware chains: public (high rate limit, any RBAC) vs admin (low rate limit, admin-only RBAC, audit trail)
   - Handler grouping: public (12 handlers), admin (5 handlers)
   - Добавлены `admin_middleware.go` (admin-specific middleware)
2. **`doc/архитектура/модули.md`** — таблица Platform internal packages: api/ обновлён с описанием 2 router'а.
3. **`doc/архитектура/пакеты-platform.md`** — ASCII-диаграмма DI графа обновлена: 2 separate routers с отдельными deps.

## Что НЕ трогать

- Файлы Go-кода handlers (admin_user.go и т.д.) — перемещение файлов произойдёт при реализации
- Nginx routes — остаются без изменений (оба router'а под одним :8080)

## Проблемы

Нет — задача чистая документация.
