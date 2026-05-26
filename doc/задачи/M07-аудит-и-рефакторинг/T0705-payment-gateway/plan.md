# T0705 — Выделение payment-gateway/ из integrations/

## План

- [x] 1. Определить границы: что уходит в payment-gateway, что остаётся в integrations
- [x] 2. Обновить `архитектура/модули.md` — новый 4-й Go сервис
- [x] 3. Добавить NATS subjects для payment-gateway
- [x] 4. Обновить `архитектура/интеграции.md` — payments separated
- [x] 5. Обновить `архитектура/безопасность.md` — PCI DSS isolation paragraph
- [x] 6. Обновить `контекст/акторы.md` — Пэйшлюз → payment-gateway
- [x] 7. Обновить Nginx routes table (платёжные endpoints через Nginx?)
- [x] 8. Создать ADR-018
- [x] 9. Обновить `архитектура/README.md` и `задачи/README.md`

## Зависимости

- Нет (T0705 выполняется первым в Wave B; T0707 зависит от T0705, но не наоборот)
