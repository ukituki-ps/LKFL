# T2205 — Нагрузочное тестирование (k6)

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Нагрузочное тестирование backend и frontend с помощью k6.
Валидация производительности при целевых нагрузках F1.

## Что сделать

### Сценарии нагрузки

- **Catalog query:** 500 RPS, P95 < 200ms, P99 < 500ms
- **Auth callback:** 100 RPS, P95 < 500ms, P99 < 1000ms
- **User profile:** 200 RPS, P95 < 100ms, P99 < 200ms
- **Combined load:** все endpoints одновременно, проверка что SLA держится

### Скрипты

- `loadtest/catalog.js` — каталог: list, filter, search, pagination
- `loadtest/auth.js` — auth: login redirect, callback, me
- `loadtest/profile.js` — user profile: get, update
- `loadtest/combined.js` — комбинированная нагрузка

### Отчёты

- HTML report: `loadtest/results/index.html`
- JSON report: `loadtest/results/results.json`
- Графики: RPS, latency (P50/P95/P99), error rate

### Требования

- k6 v0.50+
- Docker для изоляции тестов
- Результаты сохраняются в CI artifacts
- Thresholds настроены (fail if P95 exceeded)

## Критерии приёмки

- [ ] Catalog query: 500 RPS, P95 < 200ms
- [ ] Auth callback: 100 RPS, P95 < 500ms
- [ ] User profile: 200 RPS, P95 < 100ms
- [ ] Combined load scenario
- [ ] HTML + JSON reports
- [ ] Thresholds настроены (auto-fail при превышении)
- [ ] Результаты в CI artifacts
