# ADR-032: API Types — OpenAPI codegen vs ручной

**Статус:** Accepted
**Дата:** 2026-05-26
**Контекст:** M15-архитектура-фронтенда, T1504

---

## Контекст

118 API endpoints. Каждый endpoint возвращает JSON с определённой структурой. TypeScript types для API responses нужно синхронизировать с backend.

---

## Рассмотренные варианты

### Вариант А: `openapi-typescript` (codegen из OpenAPI spec)

| Плюсы | Минусы |
|-------|--------|
| Single source of truth (OpenAPI spec) | Нужно поддерживать OpenAPI spec актуальной |
| Zero drift — типы генерируются автоматически | CI step (генерация в build pipeline) |
| 118 endpoints → ~1000 строк генерируются автоматически | — |

### Вариант Б: `orval` (codegen types + hooks/fetch)

| Плюсы | Минусы |
|-------|--------|
| Генерация types + React Query hooks | Больше magic, сложнее debug |
| Full-stack type safety | Зависимость от структуры OpenAPI |
| Меньше ручного кода | — |

### Вариант В: Ручной typing

| Плюсы | Минусы |
|-------|--------|
| Полный контроль | 118 endpoints = ~1000 строк ручных types |
| Нет dependency | Drift risk: backend изменил response, фронт не знает |
| Проще в отладке | DRY: повторение структуры JSON в types |

---

## Решение

**`openapi-typescript`** — генерация types из OpenAPI spec в CI.

**Схема:**
```
openapi/spec.yaml  →  openapi-typescript  →  src/types/api.ts
```

**CI pipeline:**
1. Go backend генерирует OpenAPI spec (chi-openapi или ручное维护)
2. `openapi-typescript` генерирует `src/types/api.ts`
3. TypeScript compile-time validation

**Генерация:**
```bash
# В CI или pre-commit
openapi-typescript ./openapi/spec.yaml --output src/types/api.ts
```

---

## JSON Schema types для `AprilJsonSchemaForm`

RJSF формы используют JSON Schema для определения структуры форм — это отдельный слой от API types.

- `types/json-schema.ts` — **ручные** types для JSON Schema форм (мало форм, codegen не нужен)
- JSON Schema для admin-форм (CRUD карточек, провайдеров, правил) хранится в `types/json-schema.ts`

---

## Следствия

- `src/types/api.ts` — generated file (не коммитить вручную, CI)
- `src/types/json-schema.ts` — ручной файл для JSON Schema форм
- `.gitignore` — `src/types/api.ts` не игнорировать (коммитить сгенерированный файл для offline dev)
- CI: добавить step генерации types перед build
- Pre-commit hook: проверить что `api.ts` не изменён вручную

## Альтернативы

`orval` отклонён: генерация hooks/fetch поверх OpenAPI — избыточный уровень абстракции. React Query hooks уже покрывают data fetching (ADR-031).
