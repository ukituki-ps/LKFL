# T0705 — Выделение payment-gateway/ из integrations/

## Веха

M07-аудит-и-рефакторинг

## Контекст

Модуль `integrations/payments/` сейчас содержит:
- Пэйшлюз (онлайн-оплата Visa/MC/МИР)
- 1C (удержание из ЗП)

**Проблема:**
PCI DSS compliance требует физической изоляции платёжных данных:
- «Данные карты не хранятся на платформе. Только token от пэйшлюза» — документация
- Но модуль находится в том же сервисе, что и HR-sync, providers, webhooks
- При audit/penetration testing: auditor видит payments рядом с ПДн → риск отказа в сертификации

**Решение:**
Вынести `payments/` в отдельный сервис `payment-gateway/`:
- Собственная БД `lkfl_payments`
- Собственный NATS subscriber
- Собственный Dockerfile
- PCI DSS scoped audit

Оставить в `integrations/` только 1C (удержание из ЗП) — это БУХГАЛТЕРИЯ, не платёжная система.

### Файлы-мишени

| Действие | Файл |
|---|---|
| Новый сервис| `архитектура/модули.md` — добавить 4-й Go сервис: payment-gateway/ |
| NATS subjects | добавить `integration.payment.authorize`, `integration.payment.result` |
| Интеграции | `архитектура/интеграции.md` — payments moved out, only 1C remains |
| Безопасность | `архитектура/безопасность.md` — PCI DSS isolation |
| Архитектура README | `архитектура/README.md` — таблица содержимого |
| Акторы | `контекст/акторы.md` — Пэйшлюз interacts with payment-gateway, not integrations |
| Создать ADR | `архитектура/adr/ADR-018-payment-gateway-service.md` |
| Обновить README вехи | `задачи/README.md` — статус M07 |

### Критерии приёмки

- [x] `архитектура/модули.md` — 4 Go сервиса: platform, billing, integrations, payment-gateway
- [x] payment-gateway имеет: cmd/server/, internal/, go.mod, Dockerfile
- [x] 1C отделён от payments и оставлен в integrations (это бухгалтерия, не платёж)
- [x] NATS subjects добавлены для payment-gateway
- [x] `безопасность.md` — PCI DSS секция updated
- [x] Создан ADR-018
- [x] Файлы-мишени все перечислены выше
