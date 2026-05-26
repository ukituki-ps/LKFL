# ADR-016 — Разбиение `admin_handler.go` по бизнес-доменам

## Контекст

Текущий `admin_handler.go` в `internal/api/` делегирует в **все** бизнес-пакеты:

```go
type AdminHandler struct {
    User          *user.UserRepository
    Consent       *consent.ConsentEngine
    Engagement    *engagement.EngagementEngine  // → заменён на catalog.CatalogService + flow.FlowEngine + collections.CollectionsEngine
    Notification  *notification.NotificationEngine
    Recommendations *recommendations.RecommendationsEngine
    // ... ещё 4+ зависимостей
}
```

Этот файл:
- Содержит ~30+ методов admin-операций
- Зависит от 6+ business-пакетов
- Трудно тестируем (нужно mock'ить всё)
- Нарушает SRP: один handler выполняет CRUD для нескольких доменов

## Решение

Разделить на 4 handler-файла по бизнес-доменам:

```
internal/api/
├── admin_user.go          # /admin/users/*, /admin/periods/* → user/ + consent/
├── admin_catalog.go       # /admin/engagements/*, /admin/engagement-types/*,
#                           /admin/engagement-flows/*, /admin/collections/* → engagement/
├── admin_recommendations.go # /admin/recommendations/* → recommendations/
├── admin_analytics.go     # /admin/analytics/* → агрегация db/
└── admin_content.go       # /admin/content/*, /admin/requests/* → db/ + notification/
```

Каждый файл:
- Делегирует в 1-2 бизнес-пакета
- Имеет компактный DI (1-3 зависимости)
- Легко тестируется изолированно
- Follows SRP: один файл = один домен

## Последствия

- ✅ Читаемость: ~15 methods/файл вместо 100+ в одном файле
- ✅ Тестируемость: меньше mock'ов на тест
- ✅ Параллельная разработка: разные файлы → меньше merge conflicts
- ⚠️ Не over-engineering: это всё ещё thin handlers внутри одного Go binary (не gRPC микросервисы)

## Статус

✅ Accepted (M07, T0703)
