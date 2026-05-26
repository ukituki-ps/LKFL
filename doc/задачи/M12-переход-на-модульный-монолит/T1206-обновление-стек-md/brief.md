# T1206 — Обновление `архитектура/стек.md`

## Контекст

`стек.md` описывает NATS JetStream как обязательную технологию, 5 Redis DB, Docker compose с 14+ контейнерами. После M12 NATS → optional, Redis → key prefixes, Docker → 1 backend container.

## План

1. NATS JetStream: убрать из "обязательный" → "опционально для production split"
2. Redis: один instance, key prefixes (jwt:, asynq:, catalog:, cel:, rate:)
3. Docker compose: pg + redis + keycloak + lkfl-server + nginx = 5 контейнеров
4. Ports: убрать :8081-:8085 → :8080 + :8083
5. Observability metrics: убрать NATS-specific, добавить internal call metrics

## Ожидаемый результат

`стек.md` актуален для mono-архитектуры. NATS → optional в backend table.