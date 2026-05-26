# T1101 — Provider Gateway — отчёт

## Статус

✅ выполнено (документация)

## Что сделано

### `архитектура/модули.md` (основной артефакт):
- Переименован раздел "Integrations" → "Provider Gateway — gateway к провайдерам льгот"
- Таблица модулей сокращена с 7 до 3 (`broker/`, `providers/`, `webhook/`) — hr-sync, 1c, external удалены
- Системные адаптеры таблица очищена — таблица пуста с пояснением M11
- Синхронизация данных таблица обновлена — кадровый реестр удалён
- Admin-панель переименована "Integrations Hub" → "Provider Gateway"
- NATS JetStream секция обновлена: `integration.*` → `provider.*`, новые consumer'ы = `provider-gateway`
- Dependencies между сервисами переработаны: Platform → Provider Gateway только для engagement subjects
- Релизная политика обновлена: "Provider Gateway" вместо "Integrations"
- Все cross-links в README архитектуры обновлены

### Cross-cutting:
- `архитектура/README.md` — "Integrations Hub" → "Provider Gateway" во всех упоминаниях
- АDR-026, ADR-027, ADR-028 ссылаются на T1101 как контекст namespace rename

## Проблемы

Нет. Чистое переименование — бинарник и Docker-имя не меняются (`integrations:8082` остаётся).

## Следующие шаги

N/A — задача выполнена. T1102 выполняется последовательно.
