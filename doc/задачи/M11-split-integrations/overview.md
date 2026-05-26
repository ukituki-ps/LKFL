# M11 — Распределение модулей Integrations по доменам

## Описание

Integrations Hub нарушает SRP: 4 домена I/O (provider activation, HR sync, 1C payroll, SSO) в одном бинарнике. T0705 вынес payment-gateway, но оставил hr-sync, 1C и SSO в integrations/.

После M11: модули распределяются по 3 существующим сервисам — НЕ создаётся новый сервис.

### Что не так

| Проблема | Где | Критичность |
|--|--|--|
| hr-sync/ в integrations: HR-down → провайдеры льгот недоступны | Integrations | 🔴 Высокая |
| 1C/ рядом с providers: бухгалтерия ≠ льготные провайдеры | Integrations | 🟡 Средняя |
| NATS consumers: один сервис = один NATS consumer, падает всё | NATS | 🟡 Средняя |

### Что делается

Каждая задача M11:
1. Меняет **документацию** — архитектура модулей, пакеты, NATS registry
2. Обновляет все затронутые файлы документации для консистентности
3. Имеет чёткие файлы-мишени

### Файлы-конфликты (критически важно)

| Файл | Кто редактирует | Порядок |
|---|--|--|
| `архитектура/модули.md` | T1101 → T1102 → T1103 | Sequential — все 3 edit Integrations section |
| `архитектура/nats-subjects.md` | T1104 (единожды) | wave B — после всех T1101-T1103 |
| `архитектура/README.md` | Все | Append-safe: каждая задача добавляет 1 строку ADR |

## Волны выполнения

### Wave A (модульное перераспределение — sequential!)

```
T1101 ──► T1102 ──► T1103
```

Все 3 задачи edit `архитектура/модули.md` Integrations section. T1101 убирает hr-sync/ и 1c/ из описания → T1102 добавляет hr-sync в Platform user/ → T1103 добавляет 1C в Billing payroll.

### Wave B (после Wave A — NATS consistency)

```
T1104
```

T1104 зависит от всех трёх предыдущих — consumer mapping NATS должен знать новых consumer'ов.
**Важно:** T1104 включает решение о namespace rename `integration.engagement.*` → `provider.*` для provider-gateway subjects. Это решение documentировано в T1101 (provider-gateway identity), но выполняется в T1104.

## Веху можно закрывать когда

- [x] T1101 — Integrations описан как "Provider Gateway" (только провайдеры льгот, без hr-sync, 1C, external)
- [x] T1102 — hr-sync/ задокументирован в Platform user/
- [x] T1103 — 1C задокументирован в Billing payroll/
- [x] T1104 — NATS consumer mapping обновлён, namespace `provider.*` для provider-gateway

## Задачи вехи

| Задача | Описание | Волна | Статус |
|---|--|--|--|
| T1101 | Provider Gateway — чистый Integrations (только провайдеры льгот) | A | ✅ выполнено |
| T1102 | HR Sync → Platform user/ (HR-реестр ближе к пользователям) | A | ✅ выполнено |
| T1103 | 1C → Billing payroll (бухгалтерия ближе к финансам) | A | ✅ выполнено |
| T1104 | NATS consumer update + namespace rename `integration.*` → `provider.*` | B | ✅ выполнено |
