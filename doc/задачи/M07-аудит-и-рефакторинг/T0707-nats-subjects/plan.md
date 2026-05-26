# T0707 — Унификация NATS subjects registry (включая billing namespace)

## План

- [x] 1. Извлечь все NATS subjects из 3 источников (модули.md, billing-движок.md, engagement.md)
- [x] 2. Объединить в один master-list, устранить дубликаты и naming conflicts
- [x] 3. Определить master namespace: `billing.*` для всех billing-событий
- [x] 4. Решить: split `billing.debit` на `billing.debit.reserve` + `billing.debit.confirm`
- [x] 5. Создать `архитектура/nats-subjects.md` — таблица с полным описанием каждого subject
- [x] 6. Обновить `архитектура/модули.md` — inline tables → [link](nats-subjects.md)
- [x] 7. Обновить `архитектура/биллинг-движок.md` — inline table → [link](nats-subjects.md) + billing-events section
- [x] 8. Обновить `архитектура/engagement.md` — billing events namespace `engagement.debit.*` → `billing.*`
- [x] 9. Обновить `архитектура/интеграции.md` → link
- [x] 10. Создать ADR-020
- [x] 11. Обновить `архитектура/README.md` и `задачи/README.md`

## Зависимости

- T0705 (payment-gateway) — нужно включить новые subjects для payment-gateway
