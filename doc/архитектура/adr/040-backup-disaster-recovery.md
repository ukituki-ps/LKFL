# ADR-040: Backup и Disaster Recovery

| Поле     | Значение |
|----------|----------|
| Status   | Accepted |
| Date     | 2026-06-02 |
| Веха     | DevOps Hardening |
| Авторы   | architect-lkfl |
| Related  | ADR-030, ADR-036, ADR-039 |

## Context

LKFL хранит критические данные в нескольких хранилищах:
- **PostgreSQL 17** — основная БД (`lkfl_platform` + `keycloak` schemas) — пользователи, согласия, engagement, биллинг
- **Redis 7** — сессии, кэш, очереди Asynq — уже имеет AOF + RDB persistence
- **Keycloak** — realm data в PostgreSQL + realm config JSON в git
- **Docker volumes** — persistent data для всех сервисов

**Проблемы:**
- Нет автоматических бэкапов PostgreSQL
- Нет процедуры restore (не тестируется)
- ФСТЭК + 152-ФЗ требуют backup policy и audit log retention
- При потере данных staging — можно потерять рабочую конфигурацию и тестовые данные

## Decision

### Backup стратегия

#### PostgreSQL — основной бэкап

```bash
# Daily backup (cron на serverAi: 03:00 AM MSK)
0 3 * * * docker exec lkfl-staging-postgres-1 pg_dump -U lkfl -d lkfl_platform --format=custom --compress=6 --file=/var/lib/postgresql/data/dump_$(date +\%Y\%m\%d).dump

# Weekly full backup (включая keycloak schema)
0 4 * * 0 docker exec lkfl-staging-postgres-1 pg_dumpall -U lkfl --clean --if-exists > /backup/lkfl/full_$(date +\%Y\%m\%d).sql.gz
```

| Параметр | Значение |
|----------|----------|
| Частота | Daily (pg_dump custom format) |
| Хранение | 30 дней на сервере, 90 дней off-site |
| Формат | `.dump` (custom, compressed) для daily, `.sql.gz` для weekly |
| Размер | ~500 MB (estimated, зависит от данных) |
| Пуль | `/backup/lkfl/` на serverAi |

#### Redis — уже покрыт

- `appendonly yes` — AOF journal, fsync everysec
- `save 60 1` — RDB snapshot если ≥1 ключ за 60s
- Docker volume `staging_redis_data` — persistent на диск
- Дополнительный weekly tar volume backup

#### Keycloak

- БД данные → покрыты PostgreSQL backup
- Realm config → `infra/keycloak/realm-lkfl-sdek.json` в git
- Theme customization → `infra/keycloak/theme/` в git
- При restore: DB restore + realm reimport из git

#### Docker volumes

| Volume | Содержимое | Backup метод |
|--------|-----------|-------------|
| `staging_pg_data` | PostgreSQL data | pg_dump (не tar — надёжнее) |
| `staging_redis_data` | Redis AOF + RDB | tar backup weekly |
| Мониторинг volumes | Prometheus, Grafana, Loki | пересоздать (не критично) |

### Restore процедуры

#### PostgreSQL restore (полный)

```bash
# 1. Остановить сервер
docker compose -f docker-compose.staging.yml -p lkfl-staging down

# 2. Очистить volume
docker volume rm lkfl-staging_staging_pg_data

# 3. Поднять пустой postgres
docker compose -f docker-compose.staging.yml -p lkfl-staging up -d postgres

# 4. Восстановить из бэкапа
docker exec -i lkfl-staging-postgres-1 pg_restore -U lkfl -d lkfl_platform --clean --if-exists /backup/lkfl/dump_20260602.dump

# 5. Поднять остальные сервисы
docker compose -f docker-compose.staging.yml -p lkfl-staging up -d
```

#### PostgreSQL restore (точечный — одна таблица)

```bash
# Restore одной таблицы
docker exec -i lkfl-staging-postgres-1 pg_restore -U lkfl -d lkfl_platform --table=users --clean --if-exists /backup/lkfl/dump_20260602.dump
```

#### Redis restore

```bash
# Redis восстанавливается из AOF/RDB автоматически при перезапуске
# Если нужно восстановить из volume backup:
docker volume rm lkfl-staging_staging_redis_data
docker volume create lkfl-staging_staging_redis_data
tar xzf /backup/lkfl/redis_20260602.tar.gz -C /var/lib/docker/volumes/lkfl-staging_staging_redis_data/_data/
docker compose -f docker-compose.staging.yml -p lkfl-staging restart redis
```

### RTO / RPO

| Окружение | RTO (Recovery Time) | RPO (Recovery Point) |
|-----------|---------------------|----------------------|
| Staging | 4 часа (manual) | 1 час (daily backup) |
| Production (план) | 1 час | 5 мин (WAL archiving) |

### DR тестирование

- **Частота:** quarterly (раз в квартал)
- **Процедура:**
  1. Поднять временный postgres контейнер
  2. Восстановить из последнего бэкапа
  3. Проверить целостность данных (`\dt`, `\d+ table`, `SELECT count(*)`)
  4. Запустить smoke test
  5. Уничтожить временный контейнер
- **Результат:** записывать в `doc/архитектура/deploy-operations.md` §DR Test Log

### ФСТЭК / 152-ФЗ compliance

| Требование | Реализация |
|------------|-----------|
| Персональные данные в РФ | Сервер serverAi — физически в РФ |
| Backup ПДн | PostgreSQL backup включает таблицы с ПДн |
| Audit log retention | 90 дней minimum (Loki retention config) |
| Защита бэкапов | AES-256 encryption для off-site storage (Phase 2) |
| Доступ к бэкапам | SSH key + серверные права (ограниченный доступ) |
| Тестирование restore | Quarterly DR test |

## Consequences

### Положительные
- **Staging данные не теряются** — при crash сервера можно восстановить за 4 часа
- **ФСТЭК compliance** — backup policy + audit log retention + data localization
- **Git-backed Keycloak** — realm config восстанавливается из git
- **Redis resilience** — AOF + RDB выживает перезапуск без потери данных

### Отрицательные
- **Disk overhead:** ~2x от размера БД для 30-day retention
- **Manual restore:** требует инженерного вмешательства (нет автоматического restore)
- **Off-site storage** не реализован (Phase 2: S3/MinIO replication)

### Риски

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| Бэкап corrupt (повреждён) | Низкая | Quarterly DR test проверяет restore |
| Потеря сервера + disk | Низкая | Off-site replication (Phase 2) |
| Backup не отрабатывает (cron упал) | Средняя | Health check cron в Grafana alerting |
| RTO 4h — слишком долго для prod | — | Production: WAL archiving → RPO < 5 min |

## Production roadmap

Для production (M44+):
- **WAL archiving** → continuous backup → RPO < 5 мин
- **Point-in-time recovery** → `pg_restore` к любому моменту
- **Off-site replication** → S3/MinIO в другом DC
- **Automated backup verification** → daily checksum + test restore на временном контейнере
- **Encrypted backups** → AES-256 для off-site

## Related ADR

- **ADR-030:** CI/CD Pipeline — observability (Loki retention)
- **ADR-036:** Deploy Worker — инфраструктура серверов
- **ADR-039:** Rollback Strategy — complement к DR (rollback ≠ DR)
