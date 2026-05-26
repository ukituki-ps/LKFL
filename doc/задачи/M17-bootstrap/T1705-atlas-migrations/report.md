# T1705 — Отчёт о выполнении

## Статус

выполнено

## Что сделано

### Созданные файлы

| Файл | Описание |
|------|----------|
| `migrations/atlas.hcl` | Atlas config с external_schema для Go SDK |
| `migrations/20260526100000_init.sql` | Первая миграция: schema `lkfl_platform` + extensions (`uuid-ossp`, `pgcrypto`) |

### Миграция `20260526100000_init.sql`

```sql
CREATE SCHEMA IF NOT EXISTS lkfl_platform;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
```

## Примечания

### Atlas вместо Goose

В `brief.md` указаны Goose-аннотации (`-- +goose Up/Down`), однако инструмент миграций проекта — **Atlas** (см. `doc/архитектура/стек.md` строка 27).

Atlas использует **plain SQL** без аннотаций Goose. Формат миграций:
- Имя файла: `YYYYMMDDHHMMSS_description.sql`
- SQL-команды в порядке выполнения (Up)
- Rollback: `atlas migrate undo --count 1` (Atlas отслеживает применённые миграции в таблице `atlas_schema_migrations`)

Goose-аннотации **не добавлены** — миграция создана в чистом формате Atlas.

### Верификация

Пункты plan.yaml, требующие запущенного PostgreSQL (apply, schema check, extensions check, rollback), не могут быть проверены на текущем этапе. Требуется запущенный экземпляр PostgreSQL 17.

### Время

~15 минут
