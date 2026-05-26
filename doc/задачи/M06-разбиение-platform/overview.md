# M06 — Разбиение Platform на внутренние пакеты

## Описание

Разбиение Platform-сервиса на 3 выделенных внутренних пакета для соблюдения SRP (Single Responsibility Principle) и изоляции тестирования.

### Что не так сейчас

- `engagement/` — God Object: 5 обязанностей (каталог + eligibility + flow + collections + billing events)
- `notification-send` worker — notification-логика рассыпана по платформе без выделенного пакета
- `recommendations` — 6 endpoints без собственного пакета (М06 — отдельная бизнес-веха)

### Что делается

3 новых внутренних пакета в `internal/`:
- **`engagement/`** — каталог, eligibility engine, flow execution, collections
- **`notification/`** — шаблоны, каналы доставки (email/push/in-app), очередь
- **`recommendations/`** — правила контекст+сегмент, evaluation, debug

Количество бинарников (2), NATS subjects и Redis структура — **без изменений**.

## Веху можно закрывать когда

- [x] ADR-013 создан (почему пакеты, не сервисы)
- [x] `архитектура/пакеты-platform.md` создан (~340 строк, 7 пакетов)
- [x] `архитектура/модули.md` полностью перерисован
- [x] Спецификация и контекст согласованы
- [x] 16 критериев приёмки T0601 выполнены

## Задачи вехи

| Задача | Описание | Статус |
|---|---|---|
| T0601 | Разбиение на 3 пакета (engagement, notification, recommendations) | выполнено |
