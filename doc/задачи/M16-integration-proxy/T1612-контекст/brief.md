# T1612 — акторы.md + negative-criteria.md + настраиваемость.md

## Веха

M16-integration-proxy

## Контекст

Контекстные файлы ссылаются на `internal/integrations/` как на пакет монолита.

## Что сделать

1. `контекст/акторы.md`:
   - Администратор интеграций: управление через proxy (не mono)
   - Внешние провайдеры: взаимодействие через proxy
2. `контекст/negative-criteria.md`:
   - "Прямые вызовы platform → внешний мир" → только через Integration Proxy
3. `контекст/настраиваемость.md`:
   - Новый адаптер провайдера → `integration-proxy/adapters/` (не `integrations/providers/`)

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Обновить | `контекст/акторы.md` |
| Обновить | `контекст/negative-criteria.md` |
| Обновить | `контекст/настраиваемость.md` |

### Критерии приёмки

- [ ] акторы.md: proxy reference
- [ ] negative-criteria.md: proxy вместо direct calls
- [ ] настраиваемость.md: proxy adapters path
