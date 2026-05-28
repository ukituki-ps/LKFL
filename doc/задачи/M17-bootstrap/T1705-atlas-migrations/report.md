# T1705 — Отчёт о выполнении

## Статус

выполнено

## Что сделано

### Созданные файлы

| Файл | Описание |
|------|----------|
| `migrations/atlas.hcl` | Atlas config (без external_schema — подключается в M18+) |
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

### external_schema удалён из atlas.hcl

Оригинальная версия `atlas.hcl` содержала блок `data "external_schema" "sdk"` со ссылкой на
`./internal/schema/main.go`, который не существовал. Блок удалён — подключится в M18+
когда будет готов `internal/schema/`.

### Верификация

Пункты plan.yaml, требующие запущенного PostgreSQL (apply, schema check, extensions check, rollback),
не могут быть проверены на текущем этапе. Требуется запущенный экземпляр PostgreSQL 17.

### gen_random_uuid() compatibility function

В `brief.md` предполагалась создание compatibility function `public.gen_random_uuid()`.
Реализация её не содержит — PostgreSQL 17 имеет встроенную `gen_random_uuid()` в schema `pg_catalog`,
доступную через search_path. Добавление compatibility function избыточно для PG 17.

## Исправления (по итогам аудита)

1. **external_schema удалён** из `atlas.hcl` — блок ссылался на несуществующий `./internal/schema/main.go`.
2. **plan.yaml** — статус Goose annotations изменён с `[x]` на `[ ]` (фактически отменено, не выполнено).

### Время

~15 минут (оригинал) + ~5 минут (исправления)

## Исправления (по итогам аудита M17)

- **plan.yaml** — статус Goose annotations изменён с `[x]` на `[ ]` (фактически отменено).
  Atlas использует plain SQL без аннотаций Goose.
- **external_schema удалён** из `atlas.hcl` — блок ссылался на несуществующий файл.
