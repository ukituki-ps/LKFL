# T1702 — Отчёт

## Статус

⏳ Не начато

## Что сделано

_(пусто)_

## Проблемы

_(пусто)_

## Следующие шаги

1. shared/pkg/auth — 7 файлов (resolver, verifier, middleware, rbac, claims, errors, cache)
2. internal/auth + internal/tenant + internal/api
3. Миграции БД (tenants, tenant_brand, users) + seed data (demo tenant + users)
4. Seed Keycloak (demo realm + admin/employee users + lkfl-frontend/lkfl-server clients)
5. cmd/server/main.go — рабочий (DI, middleware, routes)
6. cmd/worker/main.go — stub (Asynq, нужен для `go build ./...`)
7. Unit-тесты (shared/pkg/auth, internal/tenant, internal/api)
