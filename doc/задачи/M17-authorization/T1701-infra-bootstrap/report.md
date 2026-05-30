# T1701 — Отчёт

## Статус

⏳ Не начато

## Что сделано

_(пусто)_

## Проблемы

_(пусто)_

## Следующие шаги

1. Инициализация go.mod (Go 1.24+)
2. cmd/server/main.go — stub
3. cmd/integration-proxy/main.go — stub
4. Dockerfile.server — stub single-stage (golang:1.24-alpine)
5. Dockerfile.proxy — stub single-stage (golang:1.24-alpine)
6. docker-compose.yml с PostgreSQL, Redis, Keycloak, Nginx
7. Makefile
8. CI smoke check (go.yml + frontend.yml)
9. Keycloak realm template + скрипт создания realm
10. Nginx конфиг ($base_domain configurable)
