# T0701 — Разбиение engagement/ на catalog/ + flow/ + collections/ + eligibility

## Веха

M07-аудит-и-рефакторинг

## Контекст

Пакет `internal/engagement/` документирован как God Object с 4 обязанностями в одном пакете:

| Подответственность | Файл (документация) | Строк API |
|---|-|--|
| Каталог (List, Get, Filter, Search, Cache) | `catalog.go` | 8 |
| Eligibility engine (AND/OR/groups evaluation) | `eligibility.go` | 3 |
| Flow execution (Activate, Complete, Revert, ExecuteStep) | `flow.go` | 4 |
| Collections management | `collections.go` | 1 |

Плюс `billing_events.go` (5 NATS событий) — пятая ответственность, которая не указана в SRP-таблице пакета.

**Проблема:**
- Eligibility engine вызывается из 3 мест: engagement-flow (свой пакет), recommendations (другой пакет), billing rule engine (третий пакет) — значит eligibility НЕ должна быть подпакетом engagement
- Flow execution зависит от billing events → это правильно держать вместе
- Catalog + collections → можно слить (collections это просто subset каталога)

**Решение — engagement/ разделён на подпакеты + eligibility вынесен:**

```
internal/
├── engagement/
│   ├── catalog/          # ← каталог + cache (EngagementType, Offer, Category)
│   ├── flow/             # ← flow execution + billing events (Flow, UserEngagement)
│   ├── collections/      # ← bundle management (EngagementCollection)
│   └── survey/           # ← Survey Engine (M13/M14)
├── eligibility/          # ← вынесен из engagement
│   ├── engine.go         # EligibilityEngine: Check, EvaluateCEL
│   └── types.go          # EligibilityResult
└── recommendations/
    └── engine.go         # уже зависит от user/ → теперь ещё от eligibility/
```

### Файлы-мишени

| Действие | Файл |
|---|---|
| Разбить engagement/ | `архитектура/пакеты-platform.md` — новый пакет `eligibility/` |
| Обновить DI граф | `архитектура/пакеты-platform.md` — eligibility зависит от user/ + pkg/ |
| Обновить таблицу пакетов | `архитектура/пакеты-platform.md` — 8 пакетов вместо 7 |
| Обновить модули | `архитектура/модули.md` — Platform: 8 internal packages |
| Обновить engagement.md | `архитектура/engagement.md` — eligibility section → ссылка на `eligibility/` пакет |
| Обновить рекомендации spec | `спецификация/api.md` — eligibility check endpoints |
| Создать ADR | `архитектура/adr/ADR-014-eligibility-extraction.md` |
| Обновить README архитектуры | `архитектура/README.md` — ADR-014 |
| Обновить README вехи | `задачи/README.md` — статус M07 |

### Критерии приёмки

- [x] `архитектура/пакеты-platform.md` — 17 internal пакетов (15 business + tenant + api)
- [x] `eligibility/` документирован как отдельный пакет с 3 публичными API
- [x] `engagement/` разделён на 4 подпакета (catalog/, flow/, collections/, survey/)
- [x] DI граф обновлён (3 пакета зависят от eligibility): flow, recommendations, billing-rule-eval
- [x] `архитектура/модули.md` — Platform: теперь 9 пакетов
- [x] Создан ADR-014 (eligibility extraction)
- [x] Файлы-мишени все перечислены выше
