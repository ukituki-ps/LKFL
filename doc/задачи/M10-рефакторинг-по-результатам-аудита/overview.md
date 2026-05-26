# M10 — Рефакторинг по результатам архитектурного аудита

## Описание

Архитектурный аудит выявил 5 конкретных точек переработки в Platform и смежных сервисах: перегруженный recommendations/ (dead weight), лишний 5-й микросервис LLM Proxy (1 agent), рассинхрон CEL Context schema между Platform/Billing, монолитный api/ HandlerDeps, дублирование auth/ в payment-gateway.

### Что не так

| Проблема | Где | Критичность |
|--|--|--|
| recommendations/: 5 файлов, 0 user journeys, нет фидбейка | Platform internal/recommendations/ | 🔴 Высокая |
| LLM Proxy: 5-й сервис для 1 agent'а, 5-я БД, не hot-path | llm-proxy/ :8085, lkfl_llm | 🟡 Средняя |
| CELContext schema: platform/cel/ vs billing/rule_engine/ — разные копии cel-go, schema drift | platform + billing | 🟡 Средняя |
| api/ HandlerDeps: 12 полей, knows о каждом домене | platform/internal/api/ | 🟡 Средняя |
| auth/ дублирован в payment-gateway/ | payment-gateway/internal/auth/ | 🟢 Низкая |

### Что делается

Каждая задача M10:
1. Меняет **документацию** (architecture docs, ADR, пакеты-platform.md)
2. Обновляет все затронутые файлы документации для консистентности
3. Имеет чёткие файлы-мишени

## Файлы-конфликты (критически важно)

Все 5 задач mass-edit `архитектура/модули.md` и `архитектура/пакеты-platform.md`.
**Правило:** внутри волны — последовательно по нумерации (T1001 → T1002 → ...).
Между волнами — строго последовательно (Wave A → Wave B).

| Файл | Кто редактирует | Порядок |
|---|--|--|
| `архитектура/пакеты-platform.md` | T1001 → T1004 → T1002 → T1003 | Wave A: T1001 → T1004, Wave B: T1002 → T1003 |
| `архитектура/модули.md` | T1005 → T1002 → T1003 → T1004 | Wave A: T1005, Wave B: T1002 → T1003 → T1004 |
| `архитектура/adr/011-monorepo.md` | T1003 → T1005 | T1005 после T1003 — T1003 создаёт shared/, T1005 расширяет |
| `архитектура/README.md` | Все | Append-safe: каждая задача добавляет 1 строку ADR |

## Волны выполнения

### Wave A (быстрые, не меняют service count)

```
T1001 ──► T1004
T1005
```

T1004 depends на T1001 (убирает Recommendations из HandlerDeps).
T1005 independent — edits модули.md, не пакеты-platform.md.

| Задача | Почему Wave A | File conflict? |
|------|-------------|-----|
| T1001 (stub recommendations) | Убирает dead package — fastest win | `пакеты-platform.md` |
| T1004 (split api) | Needs T1001 done (HandlerDeps minus Recommendations) | `пакеты-platform.md` после T1001 |
| T1005 (shared auth) | Edits `модули.md` — different file from A | `модули.md` |

### Wave B (changes service count / package structure)

```
T1002 ──► T1003
```

| Задача | Почему Wave B | Dependency |
|-----|-----|-----|
| T1002 (merge LLM Proxy) | Changes 5→4 services, mass-edits `модули.md` AND `пакеты-platform.md` | После Wave A (T1001/T1004 уже изменили пакеты-platform.md) |
| T1003 (shared CELContext) | Edits Billing rule_engine + Platform cel/ | После T1002 (T1002 убирает LLM Proxy reference из cel/ → T1003 adds shared pkg) |

## Веху можно закрывать когда

- [x] T1001 — recommendations/ стал stub-ом, убран из документации как "Phase 2"
- [x] T1002 — LLM Proxy = internal/llm/, удалён как 5-й сервис из архитектуры
- [x] T1003 — CELContext schema = shared pkg, обе стороны используют одинаковый тип
- [x] T1004 — api/ разделён на public router + admin router
- [x] T1005 — payment-gateway auth/ = ссылка на shared pkg

## Задачи вехи

| Задача | Описание | Волна | Статус |
|---|--|-|---|
| T1001 | Отложенный recommendations/ → stub | A | ✅ завершена |
| T1002 | Merge LLM Proxy → Platform internal/llm/ | B | ✅ завершена |
| T1003 | Shared CELContext pkg (platform ↔ billing) | B | ✅ завершена |
| T1004 | Split api/ → public router + admin router | A | ✅ завершена |
| T1005 | Shared auth pkg для payment-gateway | A | ✅ завершена |
