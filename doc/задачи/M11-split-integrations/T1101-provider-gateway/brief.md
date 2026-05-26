# T1101 — Provider Gateway — чистый Integrations

## Веха

M11-split-integrations

## Контекст

Сейчас `integrations/` содержит 7 модулей:
- `broker/` — NATS connection
- `providers/` — 11 адаптеров провайдеров льгот
- `hr-sync/` — кадровый реестр
- `1c/` — бухгалтерия
- `external/` — SSO Ready4/PrimeZone
- `webhook/` — incoming от провайдеров
- `admin/` — CRUD провайдеров, health dashboard, sync control

После T1102 (hr-sync → Platform) и T1103 (1C → Billing) останется:
- `broker/` — NATS connection
- `providers/` — адаптеры провайдеров
- `webhook/` — incoming
- `admin/integrations/` — CRUD
- `admin/health/` — health dashboard
- `admin/sync/` — sync control

**Решение — переименовать в provider-gateway:**
Это "чистый" Integrations: только то, что относится к провайдерам льгот.

Название service в NATS и Docker не меняется (integrations:8082 остаётся), но документация чётко разграничивает domain boundaries:

```
integrations/ → provider-gateway (conceptual rename in docs)
  ├── broker/
  ├── providers/          (11 adapters: 9 YAML + 2 hard-coded)
  ├── webhook/
  └── admin/
      ├── integrations/   (CRUD providers)
      ├── health/         (health dashboard)
      └── sync/           (sync control)
```

### Файлы-мишени

| Действие | Файл |
|-|-|--|
| Переименовать conceptually | `архитектура/модули.md` — "Provider Gateway" вместо "Integrations — единый шлюз" |
| Убрать модули | `архитектура/модули.md` — hr-sync, 1C, external → удалены (T1102, T1103, SSO → Keycloak) |
| Обновить описание | `архитектура/модули.md` — "Единственная точка контакта" → "Gateway к провайдерам льгот" |
| SSO → Keycloak | `архитектура/модули.md` — external/ удалён, SSO через Keycloak Identity Broker |
| Admin | `архитектура/модули.md` — admin теперь только провайдеры (нет HR, не 1C) |
| NATS identity | `архитектура/модули.md` — provider-gateway consumer только engagement.* subjects (Т1104 переименует namespace → `provider.*`) |

### Критерии приёмки

- [ ] `архитектура/модули.md` — Integrations переименован концептуально в "Provider Gateway"
- [ ] Таблица модулей: 7 → 4 (broker, providers, webhook, admin)
- [ ] "Единственная точка контакта с внешним миром" → "Gateway к провайдерам льгот"
- [ ] hr-sync, 1C удалены из таблицы (T1102, T1103)
- [ ] external/ удалён (SSO → Keycloak Identity Broker)
- [ ] Admin модули: только integrations, health, sync (HR sync control удалён)
- [ ] Namespace rename notice: указано что T1104 переименует `integration.*` → `provider.*`
