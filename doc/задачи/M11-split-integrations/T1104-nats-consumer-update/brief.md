# T1104 — NATS consumer update (provider-gateway + Platform + Billing)

## Веха

M11-split-integrations

## Контекст

После T1101-T1103 NATS consumer mapping разбивается на 3 сервиса вместо одного Integrations Hub:
- `provider-gateway` (ex-integrations) — consumer engagement.* subjects
- `Platform` — consumer hr.* subjects (убраны, hr-sync now internal)
- `Billing` — consumer payroll.* subjects (было integrations)

**Проблема:**
NATS registry (`nats-subjects.md`) документирует consumer mapping per subject. После split:
- `integration.payroll.submit` — consumer становится Billing (было Integrations)
- `integration.hr.pull` / `integration.hr.synced` — deleted (T1102)
- `integration.engagement.activate/deactivate/status/catalog.sync` — consumer остаётся provider-gateway

### NATS subjects до M11:

| Subject | Producer | Consumer | После M11 |
|--|--|--|--|
| `integration.engagement.activate` | platform | integrations | ✅ provider-gateway |
| `integration.engagement.deactivate` | platform | integrations | ✅ provider-gateway |
| `integration.status` | integrations | platform | ✅ provider-gateway |
| `integration.catalog.sync` | integrations | platform | ✅ provider-gateway |
| `integration.hr.pull` | platform | integrations | ❌ удалён (T1102) |
| `integration.hr.synced` | integrations | platform | ❌ удалён (T1102) |
| `integration.payroll.submit` | platform | integrations | ⚠️ billing (T1103) |

**Решение — обновить registry:**
1. Переименовать namespace: `integration.*` → `provider.*` (provider-gateway) для engagement subjects
2. Удалить `integration.hr.*` (internal call внутри Platform)
3. `integration.payroll.submit` → consumer = billing (не integrations)

### Файлы-мишени

| Действие | Файл |
|---|--|
| NATS registry | `архитектура/nats-subjects.md` — пересмотр subjects |
| Модули | `архитектура/модули.md` — NATS tables |
| Обновить billing NATS | `архитектура/модули.md` — Billing NATS consumers |
| Создать ADR | `архитектура/adr/028-nats-rename.md`

### Критерии приёмки

- [ ] `архитектура/nats-subjects.md` — subjects remapped:
  - [ ] `provider.engagement.*` → consumer = provider-gateway (ex-integrations) — namespace renamed
  - [ ] `provider.status` → producer = provider-gateway, consumer = platform — namespace renamed
  - [ ] `provider.catalog.sync` → producer = provider-gateway, consumer = platform — namespace renamed
  - [ ] `integration.hr.*` → удалены
  - [ ] `finance.payroll.submit` → consumer = billing (было `integration.payroll.submit`) — namespace renamed
- [ ] `архитектура/модули.md` — NATS tables обновлены:
  - [ ] Provider Gateway section: 4 subjects (provider.engagement.activate, provider.engagement.deactivate, provider.status, provider.catalog.sync)
  - [ ] Billing section: +1 subject (finance.payroll.submit) — Billing теперь 7 NATS subjects всего
  - [ ] Platform: no hr.* subjects
- [ ] Создан ADR-028: обоснование namespace rename (`integration.*` → `provider.*` для provider-gateway, `integration.payroll.*` → `finance.payroll.*` для billing) + consumer remapping
