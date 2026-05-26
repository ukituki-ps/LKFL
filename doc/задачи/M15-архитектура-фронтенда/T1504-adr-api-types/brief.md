# T1504 — ADR-032: API Types — OpenAPI codegen vs ручной

## Веха

M15-архитектура-фронтенда

## Контекст

118 API endpoints. Каждый endpoint возвращает JSON с определённой структурой. TypeScript types для API responses можно создавать вручную или генерировать из OpenAPI spec.

## Что решить

Как получить TypeScript types для 118 API responses:

| Вариант | Плюсы | Минусы |
|---------|-------|--------|
| **openapi-typescript** | Auto-generate, single source of truth, zero drift | Нужно поддерживать OpenAPI spec |
| **orval** | Генерация types + hooks/fetch | Больше magic, сложнее debug |
| **Ручной typing** | Полный контроль, нет dependency | 118 endpoints = ~1000 строк types, drift risk |

## Критерии

- Single source of truth (API spec → types)
- Drift prevention
- Bundle size (codegen vs manual)
- Dev experience

## Ожидаемое решение

Рекомендация: **openapi-typescript** — генерация types из OpenAPI spec в CI. `types/api.ts` — сгенерированный файл.

## Дополнительно

**JSON Schema types для `AprilJsonSchemaForm`:** RJSF формы используют JSON Schema для определения структуры форм — это отдельный типизированный слой от API types. Учет в ADR: `types/json-schema.ts` — ручные types для JSON Schema форм (мало форм, codegen не нужен).

## Результат

- `архитектура/adr/032-api-types-codegen.md` — полный ADR в формате ХАДД (вкл. JSON Schema types)
