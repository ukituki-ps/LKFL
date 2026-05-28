# Отчёт

## Статус
выполнена

## Задача
T2110 — OpenAPI Codegen для M20 endpoints

## Что сделано

### 1. Создан `frontend/openapi/spec.yaml`
- OpenAPI 3.0.3 спецификация для всех M20 endpoints
- 14 endpoints: engagements (public + admin), auth, users
- 10 schemas: EngagementCategoryResponse, EngagementOfferResponse, EngagementTypeResponse, PaginationResponse, ListResponse, UserProfile, CreateCategoryRequest, CreateTypeRequest, UpdateStatusRequest, CreateOfferRequest
- Security scheme: Bearer JWT
- Enum типы для status и type полей

### 2. Добавлен скрипт `generate-types` в package.json
- Команда: `openapi-typescript ./openapi/spec.yaml -o ./src/api/types.generated.ts`

### 3. Сгенерирован `frontend/src/api/types.generated.ts`
- openapi-typescript 7.13.0 сгенерировал 649 строк типов
- Полная типизация: paths, components/schemas, operations
- uuid.UUID → string (TypeScript native)
- *int64 → number | undefined (optional fields через `?`)
- int64 → number (required fields)
- enum → string literal unions (e.g. `"benefit" | "activity"`)

### 4. Обновлён `frontend/src/api/types.ts`
- Ре-экспорт типов из types.generated.ts с convenience aliases
- Backward compatibility: engagements.ts и user.ts работают без изменений
- Экспорт top-level типов: paths, components, operations для advanced usage

### 5. Проверка
- `tsc --noEmit` — 0 ошибок
- `npm run build` — успешный build (1.44s)

## Типы соответствуют Go struct

| Go type | TS type | Маппинг |
|---------|---------|---------|
| uuid.UUID | string | format: uuid в spec |
| string | string | direct |
| *int64 | number \| undefined | optional (`?`) |
| int64 | number | required |
| int | number | required |
| *string | string \| undefined | optional (`?`) |
| []T | T[] | array |
| enum (Go const) | `"a" \| "b"` | string literal union |

## Время
~30 минут

## Замечания
- openapi-typescript 7.13.0 сгенерировал типы автоматически без ошибок
- Spec будет заменён на автоматическую генерацию из backend (M23)
- types.ts служит адаптером между сгенерированными типами и существующим кодом
