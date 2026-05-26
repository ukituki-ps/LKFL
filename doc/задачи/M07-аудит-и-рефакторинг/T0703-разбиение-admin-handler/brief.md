# T0703 — Разбиение admin_handler.go по доменам (5 файлов)

## Веха

M07-аудит-и-рефакторинг

## Контекст

Текущий пакет `api/` документирован как "thin handlers", но `admin_handler.go` делегирует во ВСЕ бизнес-пакеты. Это нарушает принцип: один handler — один домен.

**Проблема:**
- Один файл `admin_handler.go` знает о всех 6 бизнес-пакетах
- Нет разделения ответственности между admin-endpoints разных доменов
- Тестирование невозможно: mock одного домена требует моко всех остальных
- Violation: "admin_handler.go should delegate to ONE business package"

**Что нужно сделать:**
Разделить `admin_handler.go` на 5 файлов по доменам:
1. `admin_user.go` → user + consent (HR-операции)
2. `admin_catalog.go` → engagement/catalog + collections (менеджер каталога)
3. `admin_recommendations.go` → recommendations (менеджер каталога)
4. `admin_analytics.go` → метрики, отчёты, export (администратор)
5. `admin_content.go` → content + requests (контент-менеджер)

### Файлы-мишени

| Действие | Файл |
|---|---|
| Обновить пакет api/ | `архитектура/пакеты-platform.md` — новая структура admin handlers |
| Обновить DI граф | `архитектура/пакеты-platform.md` — DI граф, middleware chain |
| Обновить таблицу handler'ов | `архитектура/пакеты-platform.md` — `admin_user_handler.go`, `admin_catalog_handler.go`, `admin_recommendations_handler.go`, `admin_analytics_handler.go` |
| Обновить API spec | `спецификация/api.md` — группировка admin endpoints по доменам |
| Обновить модули | `архитектура/модули.md` — comment о разделении admin |
| Создать ADR | `архитектура/adr/ADR-016-admin-handler-split.md` |
| Обновить README архитектуры | `архитектура/README.md` — ADR-016 |
| Обновить README вехи | `задачи/README.md` — статус M07 |

### Критерии приёмки

- [x] `архитектура/пакеты-platform.md` — 4 admin handler файла вместо 1
- [x] Каждый handler делегирует только в 1-2 бизнес-пакета
- [x] DI граф обновлён
- [x] `спецификация/api.md` — admin endpoints сгруппированы по доменам
- [x] Создан ADR-016 с обоснованием разбиения
- [x] Файлы-мишени все перечислены выше
