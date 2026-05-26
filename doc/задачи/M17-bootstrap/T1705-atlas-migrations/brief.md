# T1705 — Atlas Migrations (init)

## Веха

M17-bootstrap

## Тип

code

## Контекст

Atlas — SQL-first миграции (`doc/архитектура/стек.md` строка 27).
Первая миграция создаёт schema и extensions.

## Что сделать

### Структура

```
migrations/
├── atlas.hcl          # Atlas config
├── 20260526100000_init.sql
└── lock
```

### `atlas.hcl`

```hcl
data "external_schema" "sdk" {
  program = ["go", "run", "./internal/schema/main.go"]
}
```

### `20260526100000_init.sql`

```sql
-- +goose Up
-- +goose StatementBegin

-- Schema
CREATE SCHEMA IF NOT EXISTS lkfl_platform;

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- UUID default function (compatibility)
CREATE OR REPLACE FUNCTION public.gen_random_uuid()
RETURNS uuid
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN gen_random_uuid();
END;
$$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP FUNCTION IF EXISTS public.gen_random_uuid();
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF NOT EXISTS "uuid-ossp";
DROP SCHEMA IF EXISTS lkfl_platform CASCADE;

-- +goose StatementEnd
```

**Важно:** миграции используют **дату в имени файла** (ATM — Atlas Migration Naming), не порядковый номер.
Формат: `YYYYMMDDHHMMSS_description.sql`

## Требования

- Schema `lkfl_platform` — единственная schema для бизнес-данных
- Extensions: `uuid-ossp`, `pgcrypto` (для UUID generation)
- `gen_random_uuid()` — compatibility function (PostgreSQL 13+ имеет встроенную, но для safety)
- Goose annotations (`-- +goose Up/Down`) для rollback поддержки

## Критерии приёмки

- [ ] `migrations/` директория создана
- [ ] `20260526100000_init.sql` — schema + extensions
- [ ] Atlas migration apply работает: `atlas migrate apply --url "$DB_DSN" --dir file://migrations`
- [ ] Schema `lkfl_platform` создана в PostgreSQL
- [ ] Extensions `uuid-ossp`, `pgcrypto` установлены
- [ ] Rollback работает: `atlas migrate undo`
