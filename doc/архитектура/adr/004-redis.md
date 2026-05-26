# ADR-004: Redis для сессий и кэша

**Статус:** Accepted
**Дата:** 2026-05-22
**Контекст:** М01-создание-описания

## Контекст

Платформе нужны:
- Хранение сессий Keycloak (token cache)
- Кэширование каталога льгот (тяжёлый read, редкий write)
- Job queue для фоновых задач
- Rate limiting

## Решение

**Redis 7** с разделением по DB:

| DB | Назначение |
|----|-----------|
| 0 | JWT token cache (Keycloak session store) |
| 1 | Asynq job queue (platform workers) |
| 2 | Engagement catalog cache (TTL 6h) |
| 3 | Rate limiting counters (sliding window) |

## Альтернативы

| Вариант | Плюсы | Минусы |
|---------|-------|--------|
| Memcached | Проще | Нет persistence, нет pub/sub |
| PostgreSQL | Единая БД | Медленный кэш, нет pub/sub |
| Valkey | Fork Redis, лицензия | Меньше ecosystem |

## Следствия

- `go-redis/v9` для Go client
- AOF + RDB snapshot для persistence
- `REDIS_PASSWORD` + `requirepass` в production
