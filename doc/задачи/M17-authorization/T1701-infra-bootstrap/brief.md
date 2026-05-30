# T1701 — Инфраструктура и bootstrap проекта

## Контекст

Первый шаг реализации M17. Создаём базовую инфраструктуру: инициализируем Go-проект, настраиваем docker-compose с PostgreSQL/Redis/Keycloak/Nginx, пишем Makefile, подготавливаем Keycloak realm template и Nginx конфиг.

Это **фундамент без которого T1702 и T1703 не могут стартовать** — после этой задачи `make docker-up` поднимает полный стек, `make build` компилирует пустой проект.

**Родительский эпик:** T1700 (Полная система авторизации)
**ADR:** ADR-036 (Authorization System), ADR-003 (Keycloak), ADR-009 (Multi-tenancy)

## Что включено

- `go.mod` — инициализация модуля lkfl с ключевыми зависимостями
- `cmd/server/main.go` — stub для go build
- `cmd/integration-proxy/main.go` — stub для go build (нужен для Dockerfile.proxy в T1704)
- `Dockerfile.server` — stub single-stage (golang:1.24-alpine) для `make docker-up` в dev. Заменён на multi-stage в T1704.
- `Dockerfile.proxy` — stub single-stage (golang:1.24-alpine) для `make docker-up` в dev. Заменён на multi-stage в T1704.
- `docker-compose.yml` — PostgreSQL 17, Redis 7, Keycloak 26.x, Nginx, frontend dev
- `Makefile` — make run, build, test, migrate, docker-up, docker-down
- CI smoke check — `.github/workflows/go.yml` + `.github/workflows/frontend.yml` (заменяются на ci.yml в T1704)
- Keycloak realm template — `infra/keycloak/realm/tenant-template.json`
- Скрипт создания realm — `infra/keycloak/scripts/create-tenant-realm.sh`
- Seed demo — `infra/keycloak/seed-demo.sh` (realm demo + users admin/employee + clients)
- Nginx конфиг — `infra/nginx/lkfl.conf` ($base_domain configurable)

## Результат

- `go mod tidy` завершается без ошибок
- `go build ./...` компилирует оба бинарника (stub server + stub proxy)
- `make docker-up` поднимает PostgreSQL, Redis, Keycloak, Nginx
- Keycloak доступен, можно создать realm из template
- Nginx proxy-ит на backend/frontend
- CI smoke check работает: go build + frontend build на push/PR

## Риски

| Риск | Вероятность | Влияние | Митигация |
|------|:-----------:|:-------:|-----------|
| Keycloak 26.x — проблемы стабильности | средняя | высокое | Проверить changelog; fallback на 25.x при проблемах |
| Go 1.24 + pgx v5 — совместимость | низкая | среднее | Тестировать на раннем этапе; fallback на pgx v4 |
| Docker compose — конфликт портов в dev | высокая | низкое | Документировать порты; использовать .env для переопределения |
| Keycloak realm import — API отличается между версиями | средняя | среднее | Проверить Admin REST API v4 для target версии |
