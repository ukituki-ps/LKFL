# T1203 — Переработка `архитектура/пакеты-platform.md`

## Контекст

`пакеты-platform.md` описывает 11 внутренних пакетов Platform. После M12: 13 пакетов — добавили billing/, integrations/, payments/ (бывшие отдельные сервисы).

## План

1. Добавить `internal/billing/` — account/, transaction/, period/, rule_engine/, db/. Публичный API: Credit(), Debit(), GetBalance()
2. Добавить `internal/integrations/` — broker/, providers/, hr-sync/, 1c/, external/, webhook/. ProviderAdapter interface
3. Добавить `internal/payments/` — auth/, gateway/, api/ (thin wrappers). PCI DSS notes
4. Обновить DI граф — billing/integrations/payments доступны через interface в app/
5. Обновить Redis DB mapping → один Redis, key prefixes вместо DB 0-4

## Ожидаемый результат

`пакеты-platform.md` содержит 13 пакетов с full API spec, DI graph, и dependency table.