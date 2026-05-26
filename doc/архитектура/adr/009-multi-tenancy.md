# ADR-009: Tenant-aware архитектура (multi-tenancy)

**Статус:** Accepted
**Дата:** 2026-05-22
**Контекст:** М01-создание-описания

## Контекст

Платформа может обслуживать несколько компаний-заказчиков (СДЭК, Заказчик, future tenants). Каждая компания:
- Свой набор льгот
- Свою тему (brand)
- Свой набор провайдеров
- Свой биллинг-период

## Решение

**`tenant_id UUID` в каждой бизнес-таблице.** Keycloak realm/groups per tenant. Middleware `TenantResolver` извлекает tenant из JWT claim.

```sql
CREATE TABLE tenants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(50) UNIQUE NOT NULL,
    brand_css   TEXT NOT NULL DEFAULT '/theme/brand-default.css',
    created_at  TIMESTAMPTZ DEFAULT now()
);
```

**Изоляция данных:** все queries — с `WHERE tenant_id = $1`. Platform API — single tenant per request. Billing и Integrations — multi-tenant aware.

## Альтернативы

| Вариант | Плюсы | Минусы |
|---------|-------|--------|
| Separate DB per tenant | Полная изоляция | Дорогой scale, сложный backup |
| Row-level security (RLS) | PostgreSQL native | Сложнее debug, не все ORMs поддерживают |
| Schema per tenant | PostgreSQL native | Миграции × N schemas |

## Вердикт

**tenant_id column.** Проще всего добавить сейчас (zero cost), достаточно для ФСТЭК (все данные в одном кластере, но логически изолированы).

## Следствия

- `tenant_id` в 12+ таблиц — добавлять на этапе schema design, не после
- Keycloak claim `tenant_id` в JWT
- Brand CSS injection на основе tenant
- Future: RLS как дополнительная защита
