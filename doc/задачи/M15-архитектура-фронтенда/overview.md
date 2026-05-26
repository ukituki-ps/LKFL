# M15 — Архитектура фронтенда

## Цель

Создать полноценную архитектурную документацию фронтенда (React SPA) — аналог `модули.md` + `пакеты-platform.md` для Go-бэкенда.

## Описание

Фронтенд описан фрагментарно: есть таблица стека в `стек.md`, tree `src/` и список страниц в `модули.md`, 4 ADR (ADR-007 UI kit, ADR-008 white-label, ADR-012 Zustand, ADR-029 DS gap). Нет единого файла, который связывает всё воедино — routing, API layer, state management, компоненты, admin-страницы, performance, testing. Mobile и Forms — отдельный документ (`фронтенд-mobile-forms.md`).

**Критические проблемы, решаемые в M15:**
- ADR-029 устарел — 11 «недостающих» компонентов из которых только 2 реальные gap'ы (T1508)
- Mobile-архитектура не описана — DS `DESIGN_SYSTEM.md §8` (400+ строк нормативного текста) не интегрирован (T1509)
- `AprilProviders` не описан как корневой провайдер (T1502)
- Формы не описаны — DS поддерживает RJSF + JSON Schema (T1509)

**Контекст:** код фронтенда — 0%. Следующая веха M14 — backend-only (Survey Engine). Фронтенд-архитектура нужна до начала кодинга.

## Исходные материалы

| Артефакт | Путь |
|----------|------|
| Прототип ЛК | `артефакты/Прототип ЛК физика(1).html` |
| Структура src/ | `архитектура/модули.md` строка 215 |
| Стэк фронта | `архитектура/стек.md` строка 44 |
| ADR UI kit | `архитектура/adr/007-april-ui.md` |
| ADR white-label | `архитектура/adr/008-white-label.md` |
| ADR Zustand | `архитектура/adr/012-zustand.md` |
| ADR DS gap | `архитектура/adr/029-ds-components-gap-tz.md` |
| DS репозиторий | `DisignApril-kilo` (`@ukituki-ps/april-ui` + `@ukituki-ps/april-tokens`) |
| DS документация | `DisignApril/DESIGN_SYSTEM.md` (389 строк, §8 Mobile — 400+ строк) |
| API spec | `спецификация/api.md` (118 endpoints) |
| Акторы | `контекст/акторы.md` |

## Задачи вехи

| Задача | Описание | Тип |
|--------|----------|-----|
| T1501 | Маппинг прототип → архитектура | doc |
| T1502 | `архитектура/фронтенд.md` (основной документ) | doc |
| T1503 | ADR-031: API Data Fetching Strategy | doc |
| T1504 | ADR-032: API Types — OpenAPI codegen vs ручной | doc |
| T1505 | ADR-033: Frontend Testing Strategy | doc |
| T1506 | ADR-034: i18n — YAGNI check | doc |
| T1507 | Обновление навигации (NAVIGATION.md, README.md, модули.md) | doc |
| T1508 | Обновить ADR-029 (актуализация DS gap analysis: 11 → 2) | doc |
| T1509 | Mobile + Forms архитектура → `фронтенд-mobile-forms.md` (отдельный документ) | doc |

## Exit criteria

- [x] `архитектура/фронтенд.md` создан (10 разделов: A→J, §§I-J краткие + ссылка на `фронтенд-mobile-forms.md`)
- [x] `архитектура/фронтенд-mobile-forms.md` создан (T1509)
- [x] 4 ADR созданы и приняты (ADR-031 → ADR-034)
- [x] ADR-029 обновлён (11 → 2 реальных gap'а)
- [x] Маппинг прототипа завершён — 0 расхождений
- [x] Mobile-архитектура: `AprilMobileShellBar`, breakpoints, touch-ориентиры, жесты
- [x] Forms-архитектура: Zod + react-hook-form, wizard, survey, admin-формы
- [x] NAVIGATION.md + архитектура/README.md + модули.md обновлены
