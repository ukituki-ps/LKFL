# T1612 — Отчёт

## Веха

M16-integration-proxy

## Выполнено

Обновлены контекстные файлы для согласования с ADR-035 (Integration Proxy). Все ссылки на `internal/integrations/` как на пакет монолита заменены на Integration Proxy (`lkfl-integration-proxy`).

### Изменения по файлам

#### `контекст/акторы.md`

- **Менеджер каталога** — взаимодействие `integrations.ProviderGateway` (direct call) → `integrationclient.IntegrationClient` (gRPC → Integration Proxy)
- **Администратор интеграций** — обновлён блок M12→M16: управление через Integration Proxy, YAML config в `provider-configs/`, адаптеры в `integration-proxy/adapters/`, добавлен circuit breaker state
- **Платформа льгот** — `integrations/` (direct Go call) → `integrationclient/` (gRPC → Integration Proxy)
- **Провайдеры льгот** — `integrations.ProviderGateway.Activate/Deactivate()` (direct call) → gRPC через Integration Proxy
- **Матрица взаимодействий** — обновлены 4 строки:
  - `lkfl-server integrationclient/ → Integration Proxy` (gRPC localhost)
  - `Integration Proxy → Провайдеры льгот` (REST API direct call)
  - `Провайдеры льгот → Integration Proxy` (webhook)
  - `Integration Proxy → lkfl-server` (HTTP POST /internal/webhook/callback)
  - `Администратор интеграций → lkfl-server → Integration Proxy` (admin UI → gRPC proxy)

#### `контекст/negative-criteria.md`

- **Критерий #9** — «Прямые вызовы platform → внешний мир»: `internal/integrations/` direct Go call → Integration Proxy (gRPC, ADR-035)
- Добавлены детали: async activate/deactivate, sync status/catalog, circuit breaker per provider
- Исключения (не затронуты): HR sync `user/HRSync`, Payments `internal/payments/`, 1C payroll `billing/payroll/`

#### `контекст/настраиваемость.md`

- **Новый адаптер провайдера (SAML/SOAP/проприетарный)** — путь: `integration-proxy/adapters/` (было: `integrations/providers/`)
- **Что всегда требует кода** — ProviderAdapter уточнён: Go interface в Integration Proxy

#### `контекст/философия.md`

- **Три нуля — Нулевая привязка к льготам** — `Integrations Hub` → `Integration Proxy` (был устаревший термин M12)

### Что НЕ тронuto

- HR sync (`user/HRSync`) — оставлен direct REST pull (не затронут Integration Proxy)
- Payments (`internal/payments/`) — оставлен direct call (PCI DSS изоляция, ADR-018)
- 1C payroll (`billing/payroll/`) — оставлен direct REST

## Время

~30 минут

## Замечания

Все изменения согласованы с ADR-035. Консистентность проверена: термин «Integration Proxy» используется единообразно, пути `integration-proxy/adapters/` и `integrationclient/` соответствуют структуре из ADR-035.
