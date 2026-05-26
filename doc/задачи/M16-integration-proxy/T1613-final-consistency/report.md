# T1613 — Финальная проверка консистентности M16

## Веха

M16-integration-proxy

## Что сделано

Проведена полная проверка консистентности всей документации после M16 (Integration Proxy).

### Проверка 1: `internal/integrations/` — ✅ OK

**grep по `doc/` нашёл ~40 ссылок.** Все проверены:

- ✅ Все ссылки в историческом контексте (M12, ADR-024, ADR-005) имеют маркеры `M12:` или описывают переход
- ✅ `модули.md`, `пакеты-platform.md` — обновлены (M16 ноты: `internal/integrations/` → `integration-proxy/`)
- ✅ `api.md` — backend mapping обновлён (`integrationclient/` → gRPC → proxy)
- ⚠️ **schema.md** — найдены 3 некорректные ссылки `internal/integrations/` как Go owner:
  - `providers` (строка 988): Go owner → исправлен на `lkfl-integration-proxy`
  - `external_services` (строка 1008): Go owner → исправлен на `internal/integrationclient/`
  - Go package → Table mapping (строка 1285): исправлен на `internal/integrationclient/`

### Проверка 2: ProviderGateway — ✅ OK

**grep по `doc/` нашёл ~15 ссылок.** Все проверены:

- ✅ Все ссылки в историческом контексте (M12, ADR-024, ADR-035)
- ✅ `интеграции.md` — описывает ProviderAdapter interface (реализация в proxy)
- ✅ `пакеты-platform.md` — описывает переход `ProviderGateway.Activate()` → `IntegrationClient.Activate()` (gRPC)
- ✅ `api.md` — backend mapping: `IntegrationClient.Activate()` → gRPC → proxy
- ⚠️ Ни одной ссылки не осталось, где ProviderGateway описан как direct call из монолита

### Проверка 3: "1 бинарник" / "один бинарник" — ✅ Исправлено

**grep по `doc/` нашёл ~50 ссылок.** Проанализированы:

- ✅ Исторические контексты (ADR-024, ADR-030, ADR-005, M12 задачи) — оставлены без изменений
- ✅ `модули.md` — уже обновлён (TL;DR: "два бинарника")
- ⚠️ **акторы.md** (строки 164, 243): исправлено "1 бинарник" → "два бинарника" с M16 нотами
- ⚠️ **README.md** (строка 71): исправлено "Один бинарник" → "Два бинарника"
- ⚠️ **README.md** (строка 77): ASCII diagram обновлена (добавлен блок proxy)
- ⚠️ **README.md** (строка 82): убран `integrations/`, добавлен `integrationclient/`
- ⚠️ **README.md** (строка 109): "Бинарников | 1" → "Бинарников | 2"
- ⚠️ **README.md** (строка 110): "Internal-пакетов | 17" → "16"
- ⚠️ **README.md** (строка 112): "Интеграций" → обновлено (через Integration Proxy)
- ⚠️ **README.md** (строка 117): "Таблиц БД | 37" → "43 (37 + 6)"
- ⚠️ **README.md** (строка 135): ADR-024 описание → добавлено "Исключение: Integration Proxy (ADR-035)"

### Проверка 4: ADR-024 — ✅ Исправлено

- ✅ Добавлена секция "M16: Exception — Integration Proxy" перед "Статус"
- ✅ Статус обновлён: "✅ Accepted (M12, T1201) — с исключением M16 (Integration Proxy)"

### Проверка 5: NAVIGATION.md — ✅ OK

- ✅ `integrationclient/` строка в пакеты-platform (строка 61)
- ✅ `integration-proxy/` строка в пакеты-platform (строка 62)
- ✅ Integration Proxy навигация (строка 37)
- ✅ Критическое правило №10 (строка 216) — актуально
- ✅ ADR-035 в ADR навигации (строка 200)

### Проверка 6: api.md backend mapping — ✅ OK

- ✅ External providers → `integrationclient/` → gRPC → proxy (строка 644, 876)
- ✅ Ни одной ссылки на `integrations.ProviderGateway.Activate()` как текущей реализации

### Проверка 7: schema.md Go package → Table mapping — ✅ Исправлено

- ✅ `lkfl-integration-proxy` в списке mapping (строка 1286)
- ✅ `internal/integrations/` → `internal/integrationclient/` для `external_services` (строка 1285)
- ✅ Go owner для `providers` таблицы (строка 988): `lkfl-integration-proxy`
- ✅ Go owner для `external_services` (строка 1008): `internal/integrationclient/`

### Дополнительно

- ✅ **ADR-035** — статус обновлён: Proposed → Accepted (M16 завершён)
- ✅ **README.md** — ADR счётчик: 34 → 35 файлов, Accepted: 29 → 30

## Итого

| Проверка | Статус | Исправлено |
|----------|--------|-----------|
| internal/integrations/ | ✅ | 3 исправления в schema.md |
| ProviderGateway | ✅ | 0 (все OK) |
| "1 бинарник" | ✅ | 8 исправлений (акторы.md ×2, README.md ×6) |
| ADR-024 exception | ✅ | 1 секция добавлена |
| NAVIGATION.md | ✅ | 0 (все OK) |
| api.md backend mapping | ✅ | 0 (все OK) |
| schema.md mapping | ✅ | 3 исправления |
| ADR-035 статус | ✅ | Proposed → Accepted |

**Всего исправлений: 15.** 0 оставшихся рассинхронов.
