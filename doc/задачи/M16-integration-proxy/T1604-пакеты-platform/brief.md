# T1604 — `архитектура/пакеты-platform.md` update

## Веха

M16-integration-proxy

## Контекст

`пакеты-platform.md` описывает `internal/integrations/` как пакет монолита. Нужно:
- Убрать integrations/ из монолита
- Добавить `internal/integrationclient/` — gRPC client к proxy
- Описать структуру `integration-proxy/`

## Что сделать

1. `internal/integrations/` → удалить из пакета монолита
2. `internal/integrationclient/` → новый пакет: gRPC client, stub для test isolation
3. `integration-proxy/` → структура: adapters/, circuitbreaker/, webhook/, grpc/, config/
4. Обновить DI graph: engagement/flow/ → integrationclient/ → gRPC → proxy
5. Обновить таблицу зависимостей
6. Обновить сводную таблицу пакетов (17 → 16 в монолите + proxy)

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Обновить | `архитектура/пакеты-platform.md` |

### Критерии приёмки

- [ ] `internal/integrations/` удалён из монолита
- [ ] `internal/integrationclient/` описан (gRPC client + interface)
- [ ] `integration-proxy/` структура описана
- [ ] DI graph обновлён
- [ ] Сводная таблица: 16 пакетов в монолите + proxy
