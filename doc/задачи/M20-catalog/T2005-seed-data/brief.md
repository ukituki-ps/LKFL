# T2005 — Seed Data: Каталог

## Веха

M20-catalog

## Тип

code

## Контекст

Seed данные для каталога — категории и типы энгейджментов для tenant sdek.
Расширяем `cmd/seed/main.go` из T1805.

## Что сделать

### Категории (СДЭК референс)

| Slug | Name | Icon | Sort |
|------|------|------|------|
| dms | ДМС | Shield | 1 |
| fitness | Фитнес | Dumbbell | 2 |
| nutrition | Питание | Utensils | 3 |
| education | Образование | BookOpen | 4 |
| wellness | Благополучие | Heart | 5 |
| merchandise | Мерч | ShoppingBag | 6 |

### Типы (референс)

| Slug | Name | Category | Type | Status | Cost | Provider |
|------|------|----------|------|--------|------|----------|
| dms-federalsec | ДМС FederalSec | dms | benefit | active | 0 | FederalSec |
| fitness-worldclass | WorldClass | fitness | benefit | active | 3500 | WorldClass |
| fitness-sportmaster | Sportmaster | fitness | benefit | active | 2000 | Sportmaster |
| nutrition-yate | Ятэ | nutrition | benefit | active | 1500 | Ятэ |
| activity-nps | NPS опрос | wellness | activity | active | NULL | LKFL |
| activity-feedback | Обратная связь | wellness | activity | active | NULL | LKFL |

### Реализация

Добавить в `cmd/seed/main.go`:

```go
// Seed categories
categories := []struct{ slug, name, icon string; sort int }{
    {"dms", "ДМС", "Shield", 1},
    {"fitness", "Фитнес", "Dumbbell", 2},
    // ...
}

// Seed engagement types
types := []struct{ slug, name, categorySlug, typ, status string; cost *int64; provider string }{
    {"dms-federalsec", "ДМС FederalSec", "dms", "benefit", "active", nil, "FederalSec"},
    {"fitness-worldclass", "WorldClass", "fitness", "benefit", "active", ptrInt64(3500), "WorldClass"},
    // ...
}
```

## Требования

- 6 категорий (ДМС, фитнес, питание, образование, благополучие, мерч)
- 6 типов (3 benefit + 2 activity + 1 promo)
- Tenant: sdek (из seed)
- Idempotent (ON CONFLICT DO NOTHING)

## Критерии приёмки

- [ ] 6 категорий seeded
- [ ] 6 типов seeded
- [ ] Tenant: sdek
- [ ] Idempotent
- [ ] Каталог загружается через API
