# Отчёт M15 — Архитектура фронтенда

## Статус

✅ Завершена

## Что сделано

### Созданные документы

| Документ | Задача | Строк | Описание |
|----------|--------|-------|----------|
| `архитектура/фронтенд.md` | T1502 | ~400 | Единый документ архитектуры фронтенда (10 разделов A→J) |
| `архитектура/фронтенд-mobile-forms.md` | T1509 | ~250 | Mobile + Forms детально (AprilMobileShellBar, Zod + react-hook-form, wizard, survey) |

### Созданные ADR

| ADR | Задача | Строк | Решение |
|-----|--------|-------|---------|
| ADR-031 | T1503 | 123 | React Query (@tanstack/react-query) для data fetching |
| ADR-032 | T1504 | 84 | openapi-typescript для генерации types из OpenAPI spec |
| ADR-033 | T1505 | 126 | Vitest + RTL (unit) + Playwright (E2E) |
| ADR-034 | T1506 | 109 | i18n — YAGNI: `lib/translations/ru.ts` (не i18next) |

### Обновлённые документы

| Документ | Задача | Что изменено |
|----------|--------|-------------|
| `архитектура/adr/029-ds-components-gap-tz.md` | T1508 | 11 → 2 реальных gap'а, добавлен раздел «Компоненты LKFL vs DS» |
| `NAVIGATION.md` | T1507 | +фронтенд.md, +mobile-forms.md, +ADR-031→034, +критические правила |
| `архитектура/README.md` | T1507 | +фронтенд.md, +фронтенд-mobile-forms.md, +ADR-031→034 |
| `архитектура/adr/README.md` | T1507 | +ADR-031→034 (34 ADR) |
| `архитектура/модули.md` | T1507 | tree `src/` обновлено, +ссылки на фронтенд-документы |

## Задачи

| Задача | Статус | Описание |
|--------|--------|----------|
| T1501 | ✅ | Маппинг прототип → архитектура: 0 расхождений |
| T1502 | ✅ | `фронтенд.md` (10 разделов A→J) |
| T1503 | ✅ | ADR-031: Data Fetching → React Query |
| T1504 | ✅ | ADR-032: API Types → openapi-typescript |
| T1505 | ✅ | ADR-033: Testing → Vitest + Playwright |
| T1506 | ✅ | ADR-034: i18n → YAGNI (ru.ts) |
| T1507 | ✅ | Навигация (NAVIGATION.md, README.md, модули.md) |
| T1508 | ✅ | ADR-029 обновлён (11 → 2 gap'а) |
| T1509 | ✅ | `фронтенд-mobile-forms.md` (Mobile + Forms) |

## Exit criteria — все выполнены

- [x] `архитектура/фронтенд.md` создан (10 разделов: A→J, §§I-J краткие + ссылка на `фронтенд-mobile-forms.md`)
- [x] `архитектура/фронтенд-mobile-forms.md` создан (T1509)
- [x] 4 ADR созданы и приняты (ADR-031 → ADR-034)
- [x] ADR-029 обновлён (11 → 2 реальных gap'а)
- [x] Маппинг прототипа завершён — 0 расхождений
- [x] Mobile-архитектура: `AprilMobileShellBar`, breakpoints, touch-ориентиры, жесты
- [x] Forms-архитектура: Zod + react-hook-form, wizard, survey, admin-формы
- [x] NAVIGATION.md + архитектура/README.md + модули.md обновлены
