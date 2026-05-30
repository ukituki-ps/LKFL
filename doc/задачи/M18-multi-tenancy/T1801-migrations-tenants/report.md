# T1801 — Отчёт о выполнении

## Статус

✅ выполнено

## Что сделано

### `migrations/20260526110000_tenants.sql`

Создана миграция Atlas (plain SQL, без goose аннотаций):

- **`lkfl_platform.tenants`** — таблица tenants:
  - `id` UUID PK DEFAULT gen_random_uuid()
  - `slug` VARCHAR(64) NOT NULL UNIQUE
  - `name` VARCHAR(255) NOT NULL
  - `status` VARCHAR(20) CHECK (active, suspended, terminated)
  - `settings` JSONB DEFAULT '{}'
  - `created_at`, `updated_at` TIMESTAMPTZ
  - Index `idx_tenants_slug` на slug

- **`lkfl_platform.tenant_brand_config`** — white-label брендирование:
  - `id` UUID PK DEFAULT gen_random_uuid()
  - `tenant_id` UUID FK → tenants(id) ON DELETE CASCADE
  - `primary_color`, `secondary_color` VARCHAR(7)
  - `logo_url`, `favicon_url` TEXT
  - `brand_name` VARCHAR(255)
  - `css_variables` JSONB DEFAULT '{}'
  - `meta_title`, `meta_description`
  - `created_at`, `updated_at` TIMESTAMPTZ
  - Unique index `idx_brand_config_tenant` на tenant_id

## Критерии приёмки

- [x] `tenants` таблица создана
- [x] `tenant_brand_config` таблица создана
- [x] UNIQUE index на `slug`
- [x] CHECK constraint на `status`
- [x] UNIQUE index на `tenant_brand_config.tenant_id`
- [x] ON DELETE CASCADE работает
- [ ] Migration apply + rollback OK (требует БД)

## Время

~10 мин
