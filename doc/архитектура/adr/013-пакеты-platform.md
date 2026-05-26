# ADR-013: Platform — internal пакеты вместо gRPC микросервисов

**Статус:** Accepted
**Дата:** 2026-05-24
**Контекст:** M06-разбиение-platform, T0601

---

> **M12:** 3 новых пакета (billing/, integrations/, payments/) → один бинарник. Теперь 13 internal пакетов вместо 4 сервисов. [ADR-024](./024-modular-monolith.md)

---

## Контекст

После M05 (унификация Engagement) модуль `engagement/` внутри Platform-сервиса стал God Object с 5 обязанностями одновременно:
1. **Каталог энгейджментов** — CRUD, фильтры (type=benefit/activity), поиск, кэширование
2. **Eligibility engine** — AND/OR/groups evaluation, самая сложная бизнес-логика платформы
3. **Flow execution** — activation/completion/revert, 3 Asynq worker'а
4. **Collections** — bundle management (~40% поверхности Platform)
5. **Billing event publishing** — **M12:** direct call `billing.BillingService.Credit(ctx, ...)` (был NATS)

Помимо этого:
- **Notification-логика** рассыпана между Asynq worker `notification-send` и api-handlers без выделенного пакета. Нет `Channel` interface, нет template engine.
- **Recommendations** (рекомендательная система) — 5 admin endpoints + 1 user endpoint + 1 debug endpoint, всё перемешано в `api/` без изоляции для тестирования.

**Проблемы:**
- **SRP violation:** `engagement/` нарушает Single Responsibility Principle, прямо противореча принципы SOLID проекта
- **Нет изоляции для тестирования:** eligibility engine (AND/OR/groups/segments) — не возможно написать unit-тесты без запуска Platform целиком
- **Notification scattered:** логика уведомлений между worker и handlers без централизации
- **Recommendations невидимы:** бизнес-логика рекомендаций не имеет собственного пакета

---

## Решение

Разделить Platform на **3 новых internal пакета** (не сервисы, не отдельные бинарники):

| Пакет | Назначение | Что переезжает |
|-------|----------|-------------|
| `internal/engagement/` | Каталог, flow, collections, survey | God Object `engagement/` → 4 подпакета (catalog/, flow/, collections/, survey/) + eligibility вынесен |
| `internal/notification/` | Шаблоны, каналы (email/push/in-app), очередь | Worker `notification-send` + api-handlers |
| `internal/recommendations/` | Правила контекст+сегмент, evaluation, debug | Admin endpoints `/admin/recommendations/*` + `/recommendations` + debug endpoint |

Оставшиеся пакеты без изменений:
- `internal/auth/` — OIDC verification, middleware, RBAC
- `internal/user/` — CRUD пользователей
- `internal/consent/` — lifecycle ПДн
- `internal/api/` — thin HTTP handlers → делегируют в business-пакеты

---

## Альтернативы

| Вариант | Описание | Плюсы | Минусы | Вердикт |
|---------|---------|-------|--------|--------|
| **internal пакеты** (выбрано) | `internal/engagement/`, `internal/notification/`, `internal/recommendations/` | Test isolation без запуска Platform, SRP, 0 изменений в infra | Код-миграция, временная сложность рефакторинга | ✅ |
| **gRPC микросервисы** | engagement-svc, notification-svc, recommendations-svc | Complete isolation, independent deploy | 5+ бинарников, gRPC clients, distributed transactions, 5+ Nginx upstream | ❌ overkill |
| **Оставить как есть** | God Object `engagement/` + scattered notification/recommendations | Ничего не менять | SRP violation, no test isolation, growing maintenance cost | ❌ технический долг |
| **Feature modules** | engagement, notification, recommendations как Go module'ы в subdirectory'х | Package isolation | Same complexity как internal/, extra go.mod management, no benefit | ❌ unnecessary complexity |

---

## Аргументы «за»

1. **Test isolation:** eligibility engine тестится без Platform; Channel interface тестится с mock; recommendations evaluator независим
2. **SRP compliance:** каждый пакет — одна обязанность (catalog, delivery engine, recommendation engine)
3. **Readability:** разработчик нового пакета видит < 2000 строк вместо > 6000 строк God Object'а
4. **Zero infra cost:** бинарников по-прежнему 2, NATS subjects без изменений, Redis без изменений, Nginx без изменений
5. **DI injection:** каждый бизнес-пакет принимает интерфейсы зависимостей → unit-тесты без реальных зависимостей

---

## Аргументы «против»

1. **Код-миграция:** перемещение ~3000 строк кода между пакетами
2. **Временная сложность:** во время миграции возможны рассогласования между документацией и кодом
3. **Refactoring риск:** rename package + update imports + update tests — чек-лист из 16 критериев приёмки

---

## Вердикт

**internal пакеты.** Код-миграция — единственный правильный выбор при текущем масштабе (131 endpoint, 1 команда). gRPC-микросервисы — overkill: Platform остаётся API-агрегатором без отдельного SLA для каждого business-домена.

Порог перехода на mикросервисы:
- Пакет > 5000 строк кода
- SLA отличается (99.99% vs 99.9%)
- > 3 команды параллельно работают на одном пакете
- 300+ endpoints в Platform

Ни один из 3 новых пакетов не соответствует этим критериям сейчас.

---

## Следствия

1. **Документация:** `пакеты-platform.md` (~300 строк) — детальные спецификации всех 13 пакетов (**M12:** +billing/, integrations/, payments/)
2. **`модули.md`:** полная перерисовка секции Platform → lkfl-server с 13 internal пакетами (**M12**)
3. **Спецификация:** endpoints api.md распределены по бизнес-пакетам
4. **API контракты не меняются:** HTTP endpoints остаются теми же (134 шт) — меняется только внутренняя маршрутизация
5. **User-facing поведение не меняется:** Frontend не видит изменения (тот же Nginx route `/api/`)
6. **Asynq workers mapping:** каждый worker делегирует в конкретный пакет (notification-send → notification/, engagement-* → engagement/)
7. **Следующая код-миграция:** отдельная веха после M06
8. **M12:** [ADR-024](./024-modular-monolith.md) — modular monolith: один бинарник, 13 internal пакетов, Go interfaces вместо NATS
