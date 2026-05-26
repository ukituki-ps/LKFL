# ADR-018 — Выделение `payment-gateway/` из `integrations/` → `internal/payments/` после M12

**Статус:** ⚠️ Note: M12 merged into `internal/payments/`
**Дата:** обновлено 2026-05-25 (M12)

> **M12:** `payment-gateway/` слит в lkfl-server как `internal/payments/` ([ADR-024](./024-modular-monolith.md)). PCI DSS isolation через separation of concerns на уровне кода: separate credentials в Vault, token-only, TLS 1.3. NATS subjects `payment.*` → direct Go call `payments.PaymentGateway.*()`. `/payments/v1/` → `/api/v1/payments/`.

## Контекст (исторический)

Модуль `payments/` внутри `integrations/` выполняет:
- Авторизацию платежа (Visa/MC/МИР)
- Подтверждение (capture)
- Возврат (capture + payroll_deduction)
- Передачу заявления на удержание из ЗП в 1С

`payments/` сейчас обрабатывает карточные данные (хоть и token-only). PCI DSS требует физической изоляции платёжной инфраструктуры.

## Решение (историческое)

Выделить `payment-gateway/` как 4-й независимый Go-сервис:

```
lkfl/
├── platform/            # Go API + Worker (существующий)
├── billing/             # Go API (существующий)
├── integrations/        # Go API (существующий, payments/ убран)
├── payment-gateway/     # ← НОВЫЙ (T0705) — **M12: слит в internal/payments/**
│   ├── cmd/server/
│   ├── internal/
│   │   ├── auth/        # JWT validation (same as platform)
│   │   ├── gateway/     # payment processing logic
│   │   └── api/         # thin REST handlers
│   ├── go.mod
│   └── Dockerfile
```

### Почему отдельный сервис, а не модуль (исторические аргументы):

| Критерий | Платёж-шлюз | M12 статус |
|-----|----------|------|
| **PCI DSS** | Требует физической изоляции | ✅ Separation of concerns: separate credentials, token-only, TLS 1.3 |
| **Режим работы** | Request-response, не event-driven | ✅ В mono-режиме — direct Go call, не NATS |
| **Релиз** | Независимый | ❌ Один бинарник, один релиз |
| **Scaling** | Отдельный | ❌ Масштабируется вместе с lkfl-server |

### NATS subjects (исторические):

| Subject | Producer | Consumer | Описание |
|-----|-----|---|---|--|
| `payment.authorize` | platform | payment-gateway | Авторизация платежа |
| `payment.result` | payment-gateway | platform | Результат авторизации |
| `payment.capture` | platform | payment-gateway | Подтверждение платежа |
| `payment.payroll.submit` | platform | payment-gateway | Передача заявления на удержание ЗП |

> **M12:** Эти subject'ы больше не используются — заменили direct Go call.

## Последствия

- ✅ PCI DSS compliance: separation of concerns maintained via code-level isolation
- ✅ Интеграции Hub: только агрегация провайдеров льгот и HR-синхронизация
- ⚠️ Count Go services: 3 → 4 → **M12: 1 бинарник**
- ⚠️ NATS subjects +4 → **M12: 0 NATS subjects**

## Статус

⚠️ Note: M12 merged into `internal/payments/`. См. [ADR-024](./024-modular-monolith.md).
