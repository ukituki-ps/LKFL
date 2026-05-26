# T1608 — `архитектура/стек.md` — +gRPC, 2 binaries

## Веха

M16-integration-proxy

## Контекст

`стек.md` описывает технологии проекта. Нужно добавить gRPC и обновить описание бинарников.

## Что сделать

1. Backend: 1 бинарник → 2 бинарника (один go.mod)
2. Communication: Go interfaces + gRPC (localhost)
3. Dependencies: +google.golang.org/grpc, +google.golang.org/protobuf
4. Metrics: +integration_* метрики (6 новых)
5. Docker: 5 → 6 контейнеров

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Обновить | `архитектура/стек.md` |

### Критерии приёмки

- [ ] Backend: 2 бинарника, 1 go.mod
- [ ] Communication: Go interfaces + gRPC
- [ ] Dependencies: +grpc, +protobuf
- [ ] Metrics: +6 integration_* метрик
- [ ] Docker: 6 контейнеров
