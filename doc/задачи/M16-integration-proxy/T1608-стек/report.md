# T1608 — Отчёт: обновление `архитектура/стек.md`

## Веха
M16-integration-proxy

## Дата
2026-05-26

## Что сделано

### 1. Краткое описание
- Обновлена ссылка на ADR: `ADR 001–012` → `ADR 001–035`
- Добавлен абзац о бинарниках: 2 бинарника (`lkfl-server` + `lkfl-integration-proxy`), 1 `go.mod`
- Указана межмодульная коммуникация: Go interfaces + gRPC (localhost)

### 2. Backend зависимости
- Добавлена строка `gRPC` — `google.golang.org/grpc` 1.71.0 с обоснованием ADR-035
- Добавлена строка `Protobuf` — `google.golang.org/protobuf` 1.36.0 с обоснованием ADR-035

### 3. Межмодульная коммуникация
- Backend: `Message Broker` → `Межмодульная коммуникация`: Go interfaces + gRPC, с пояснением ADR-035
- Инфраструктура: `Message Broker` → `Межсервисная коммуникация`: gRPC (localhost), lkfl-server ↔ lkfl-integration-proxy

### 4. Метрики (Prometheus)
Обновлены 2 существующие + добавлены 4 новые метрики (все с маркировкой **ADR-035**):

| Метрика | Тип | Статус |
|---------|-----|--------|
| `integration_provider_latency_seconds` | Histogram | ✅ обновлено описание |
| `integration_provider_errors_total` | Counter | ✅ обновлено описание |
| `integration_circuit_breaker_state` | Gauge | ✅ новая |
| `integration_sync_duration_seconds` | Histogram | ✅ новая |
| `integration_webhook_total` | Counter | ✅ новая |
| `integration_dead_letters_total` | Counter | ✅ новая |

### 5. Docker Compose
- Обновлена строка Orchestration: 6 контейнеров (lkfl-server, lkfl-integration-proxy, postgres, redis, keycloak, nginx)

## Изменённые файлы

| Файл | Изменения |
|------|-----------|
| `doc/архитектура/стек.md` | +8 строк, изменён формат таблиц Backend/Инфраструктура/Метрики |
| `doc/задачи/M16-integration-proxy/T1608-стек/plan.yaml` | progress: 0% → 100%, все чекбоксы [x] |
| `doc/задачи/M16-integration-proxy/T1608-стек/report.md` | создан |

## Консистентность с ADR-035

Все изменения согласованы с ADR-035:
- Бинарники: 2 бинарника, 1 go.mod ✅
- gRPC localhost для mono ↔ proxy ✅
- Circuit breaker states (0/1/2) ✅
- 6 integration_* метрик полностью совпадают с таблицей ADR-035 ✅
- Docker контейнеры: proxy добавлен ✅

## Время
~15 минут

## Замечания
Нет. Все критерии приёмки выполнены.
