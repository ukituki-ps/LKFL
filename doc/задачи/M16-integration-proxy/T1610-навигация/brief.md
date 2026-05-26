# T1610 — NAVIGATION.md + README.md update

## Веха

M16-integration-proxy

## Контекст

NAVIGATION.md и архитектура/README.md нужно обновить для навигации к Integration Proxy.

## Что сделать

1. NAVIGATION.md:
   - +строка: "Integration Proxy архитектура" → `архитектура/интеграции.md` §Proxy
   - +строка в пакеты-platform: `integrationclient/` строка
   - +критическое правило: "Прямые HTTP calls из монолита запрещены — только через proxy"
2. архитектура/README.md:
   - +Integration Proxy в описание модулей
   - +ADR-035 в таблицу ADR

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Обновить | `NAVIGATION.md` |
| Обновить | `архитектура/README.md` |

### Критерии приёмки

- [ ] NAVIGATION.md: +Integration Proxy навигация
- [ ] NAVIGATION.md: +критическое правило
- [ ] архитектура/README.md: +proxy описание
- [ ] архитектура/README.md: +ADR-035
