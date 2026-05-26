# T1607 — `архитектура/модули.md` — +proxy binary

## Веха

M16-integration-proxy

## Контекст

`модули.md` описывает один бинарник `lkfl-server`. Нужно добавить `lkfl-integration-proxy`.

## Что сделать

1. TL;DR: 1 бинарник → 2 бинарника (mono + proxy)
2. Diagram: добавить proxy между монолитом и внешним миром
3. `internal/integrations/` → перенесён в `integration-proxy/adapters/`
4. Межмодульная коммуникация: mono → proxy через gRPC (не Go interface)
5. Port mapping: 8080 (mono) + 8090 (proxy gRPC) + 8091 (proxy HTTP webhooks)
6. Docker compose: 5 → 6 контейнеров

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Обновить | `архитектура/модули.md` |

### Критерии приёмки

- [ ] TL;DR: 2 бинарника
- [ ] Diagram обновлена
- [ ] integrations/ → proxy
- [ ] gRPC communication описана
- [ ] Port mapping обновлён
- [ ] Docker compose: 6 контейнеров
