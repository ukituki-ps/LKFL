# T1001 — Отложенный recommendations/ → stub

## Веха

M10-рефакторинг-по-результатам-аудита

## Контекст

Пакет `internal/recommendations/` задокументирован как двухслойная рекомендательная система:
- engine.go — Recommend, Debug (6 публичных API)
- rules.go — CRUD правил
- evaluation.go — segment matching, context scoring
- debug.go — Trace по userId
- storage.go — persistence + hit_count + conversion_rate

**Проблема:**
1. 0 user journeys зависят от `recommend/` endpoint. Dashboard, Catalog, Points — нигде не есть рекомендации.
2. Нет механизма обратной связи: кликнул → update hit_count → re-weight. conversion_rate = 0 всегда.
3. 5 файлов + Redis cache + PostgreSQL table — overhead без ценности.

**Решение — stub-реализация:**
Заменить `RecommendationsEngine` на stub, возвращающий пустой slice. Document package как "Phase 2 — requires frontend integration first".

```go
type RecommendationsEngine struct {
    logger Logger
}

func (r *RecommendationsEngine) Recommend(ctx context.Context, userId uuid.UUID, context *EngagementContext) ([]Recommendation, error) {
    return []Recommendation{}, nil
}

func (r *RecommendationsEngine) Debug(ctx context.Context, userId uuid.UUID) (*DebugResult, error) {
    return &DebugResult{
        ContextRules: nil,
        SegmentRules: nil,
        Recommendations: nil,
    }, nil
}

// CRUD — TODO stub
```

### Файлы-мишени

| Действие | Файл |
|---|---|
| Убрать recommendations/ как full package | `архитектура/пакеты-platform.md` — отметить как stub (Phase 2) |
| Уменьшить число пакетов | `архитектура/модули.md` — Platform 11 → 10 (1 stub) |
| Убрать из DI графа | `архитектура/пакеты-platform.md` — HandlerDeps: убрать Recommendations поле |
| Убрать handler | `архитектура/пакеты-platform.md` — api/recommendation_handler.go → stub handler |
| Убрать из Asynq mapping | `архитектура/модули.md` — workers не используют recommendations |
| Обновить README архитектуры | `архитектура/README.md` — ссылки на recommendations |
| Обновить матрицу настраиваемости | `контекст/настраиваемость.md` — рекомендации → Phase 2 |

### Критерии приёмки

- [ ] `архитектура/пакеты-platform.md` — recommendations/ отмечен как stub (Phase 2, frontend not ready)
- [ ] Рекомендация engine описан как 2 метода: Recommend() → empty, Debug() → nil
- [ ] CRUD методы задокументированы как "Phase 2 — TODO"
- [ ] HandlerDeps уменьшен (убрано поле Recommendations из DI struct) в `пакеты-platform.md`
- [ ] `архитектура/модули.md` — Platform 11 пакетов → 10 (recommendations = stub)
- [ ] Redis optional cache удалён из зависимостей recommendations/ в `пакеты-platform.md`
- [ ] `контекст/настраиваемость.md` — recommendation rules → Phase 2 (frontend ready required)
- [ ] `архитектура/README.md` — ссылки на recommendations/ обновлены (stub notice)
