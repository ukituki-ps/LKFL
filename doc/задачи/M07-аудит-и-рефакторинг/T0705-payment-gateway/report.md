# T0705 — Выделение payment-gateway/ из integrations/ — отчёт

## Статус

✅ выполнено

## Что сделано

- `архитектура/модули.md` — добавлен 4-й Go-сервис: payment-gateway/
  - Структура: cmd/server/, internal/{auth, gateway, api}, go.mod, Dockerfile
  - NATS subjects: payment.authorize, payment.result, payment.capture, payment.payroll.submit
  - Порты: :8084 (HTTP API, через Nginx)
  - Управленческий API: 4 endpoints (transactions list/get/void, refund)
- 1С отделён от пэйшлюза и оставлен в integrations (бухгалтерия, не PCI DSS)
- `архитектура/интеграции.md` — payments moved out, only 1C remains (обновлены таблицы системных адаптеров)
- `архитектура/безопасность.md` — PCI DSS isolation: payment-gateway/ (отдельный сервис)
- `контекст/акторы.md` — Пэйшлюз interacts with payment-gateway, not integrations
- Создан ADR-018: обоснование отдельного сервиса (PCI DSS, request-response, independent release, scaling)
- `архитектура/README.md` — ADR-018 добавлен в таблицу
- `задачи/README.md` — статус M07 обновлён

## Проблемы

- Были stale references: `payments/` показан в секции integrations модули.md (стр.191, 235-236) — исправлено в ходе аудита
