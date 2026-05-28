# T2005 — Отчёт

## Статус

✅ Выполнено

## Что сделано

### backend/cmd/seed/main.go

Добавлены три функции и обновлён `main()`:

1. **`seedCategories(conn, tenantID)`** — загружает 6 категорий каталога:
   - dms (ДМС, Shield, sort=1)
   - fitness (Фитнес, Dumbbell, sort=2)
   - nutrition (Питание, Utensils, sort=3)
   - education (Образование, BookOpen, sort=4)
   - wellness (Благополучие, Heart, sort=5)
   - merchandise (Мерч, ShoppingBag, sort=6)

2. **`seedEngagementTypes(conn, tenantID)`** — загружает 6 типов энгейджментов:
   - dms-federalsec (ДМС FederalSec, benefit, cost=0, provider=FederalSec)
   - fitness-worldclass (WorldClass, benefit, cost=3500, provider=WorldClass)
   - fitness-sportmaster (Sportmaster, benefit, cost=2000, provider=Sportmaster)
   - nutrition-yate (Ятэ, benefit, cost=1500, provider=Ятэ)
   - activity-nps (NPS опрос, activity, cost=NULL, provider=LKFL)
   - activity-feedback (Обратная связь, activity, cost=NULL, provider=LKFL)

3. **`int64Ptr(v int64) *int64`** — вспомогательная функция для указателей на int64.

4. **`main()`** — добавлены вызовы `seedCategories` и `seedEngagementTypes` после `upsertBrandConfig`.

### Критерии приёмки

| Критерий | Статус |
|----------|--------|
| 6 категорий seeded | ✅ |
| 6 типов seeded | ✅ |
| Tenant: sdek | ✅ |
| Idempotent (ON CONFLICT DO NOTHING) | ✅ |
| go build ./cmd/seed/ компилируется | ✅ |

### Замечания

- Существующий seed (tenant + brand config) не изменён
- activity типы имеют `cost_cents = NULL` (nil *int64)
- benefit типы имеют `cost_cents` как *int64 с конкретным значением
- Все INSERT используют `ON CONFLICT (tenant_id, slug) DO NOTHING` для идемпотентности
