# T1704 — Отчёт

## Статус

⏳ Не начато

## Что сделано

_(пусто)_

## Проблемы

_(пусто)_

## Следующие шаги

1. GitHub Actions workflows (ci.yml, cd.yml, deploy.yml) — заменяют go.yml/frontend.yml из T1701
2. Dockerfiles (server, proxy, frontend) — multi-stage (заменяют stub single-stage из T1701)
3. docker-compose.prod.yml
4. testcontainers-go (PG + Redis + Keycloak)
5. Integration tests (*_integration_test.go)
6. OpenAPI spec (openapi/openapi.yaml + redocly lint в CI)
