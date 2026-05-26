# ADR-023 — Система геймификации на базе CEL + immutable фактов присвоения

## Контекст

Платформы гибких льгот не хватает слоя геймификации. Прототип показывает прогресс-бары, badge-статусы, баллы за активности — но backend-механики присвоения ачивок и уровней лояльности не спроектированы.

**Связь с существующей архитектурой:**
- ADR-021 (CEL Engine) — CEL уже есть, используется для billing, eligibility, flow, recommendations
- Engagement (`engagement.md`) — UserEngagement.completed → естественный триггер для проверки ачивок
- Пользователи (Keycloak + user/) — источник профиля для evaluation

**Проблема:** где хранить условия присвоения ачивок и как оценивать?
- Вариант A: Tag Engine (отдельный механизм условий)
- Вариант B: CEL + immutable facts (использовать единый CEL + хранить только факты присвоения)

### Вариант A — Tag Engine

Ввести отдельный движок тегов/условий специфичный для геймификации:
- Собственный YAML/JSON формат условий типа `{"completions": {"gte": 5}}`
- Собственный evaluator, независимый от CEL
- Отдельный UI-constructor в админке

**Плюсы:**
- Простой синтаксис для простых условий
- Не зависит от CEL schema evolution

**Минусы:**
- 5-й независимый механизм условий (нарушает принцип ADR-021)
- Дублирование логики evaluation
- HR-менеджер учит 5-й синтаксис
- Нельзя переиспользовать условия между доменами
- Никакой LLM-генерации для геймификации

### Вариант B — CEL + immutable facts (ВЫБРАН)

Использовать существующий CEL Engine из ADR-021 для оценки условий присвоения.
Геймификация становится 5-м доменом CEL (после billing, eligibility, flow, recommendations).

Факты присвоения (какой пользователь получил какую ачивку, когда вошёл в какой уровень) — immutable records в БД.

**Плюсы:**
- Единый синтаксис условий (ADR-021)
- LLM генерация CEL из русского текста уже работает
- Нет 5-го движка — реиспользуем `cel/` пакет
- Прозрачность: HR видит CEL expression + источник на русском
- Обратная связь: геймификация может влиять на eligibility
  (`user.achievements.contains('survey_master')`)

**Минусы:**
- Нужно расширить CELContext полями геймификации (`game.*`)
- Добавляем новый internal пакет `gamification/`

## Решение

**Вариант B: CEL + immutable facts.**

### Архитектура

```
            CEL Engine (условия)              Геймификация (факты)
┌──────────────────────────────┐      ┌─────────────────────────────┐
│ achievements.cel_condition:  │      │ achievement_grants:         │
│ "completions >= 5 &&         │──→   │  user_id,                  │
│  avg_score >= 7"             │      │  achievement_key,           │
│                              │      │  awarded_at, visible        │
│ loyalty.cel_condition:       │      └─────────────────────────────┘
│ "engagement_count >= 10"     │         ↑
└──────────────────────────────┘    CheckAchievementEngine
                                       (trigger: engagement events + cron)
```

### Ключевые решения

1. **CEL — 5-й домен оценки:** условия присвоения ачивок и уровней лояльности выражаются через CEL expression. Те же `CELGenerator`, `CELEvaluator`, `CELValidator` из ADR-021.

2. **CELContext расширение:** добавить вложенный блок `game.*` с полями:
   - `game.achievements` — ключи имеющихся ачивок (список)
   - `game.achievement_count` — количество ачивок
   - `game.engagement_count` — всего завершённых энгейджментов
   - `game.engagement_by_category` — map: категория → количество
   - `game.benefit_categories_count` — кол-во категорий льгот
   - `game.loyalty_level` — текущий уровень лояльности
   - `game.loyalty_points` — cumulative engagement points
   - `game.days_since_active` — дней с последней активности
   - `game.enps_submitted` — submit'нул ENPS
   - `game.has_family` — есть родственники в системе ДМС

3. **Immutable grants:** `achievement_grants` — never UPDATE на существующие записи. Если ачивка "отзывается" — новая запись с флагом revoked.

4. **Historичные уровни:** `user_loyalty_levels` — valid_to pattern. Один активный уровень (exited_at IS NULL), полная история переходов.

5. **Триггеры присвоения — Go callback, не NATS:**
   Gamification получает события внутри Platform, не подписываясь на NATS billing subjects.
   `FlowEngine.Complete()` → Go-callback → `TriggerHandler.OnEngagementCompleted()`
   Причины: нет race-condition, нет отдельного subscriber, simpler.

6. **Массовое присвоение — XLSX import:** паттерн из `user/registry-import-xlsx`.
   Два шаблона: бейджи + уровни. Валидация → preview → apply. Собственная таблица `gamification_import_jobs`.

7. **Новый пакет `gamification/` — 11-й внутри Platform:**
   SRP: достижения, уровни лояльности, триггеры, массовый импорт.
   Зависит от: `cel/` (evaluation), `db/` (PostgreSQL), `user/` (профиль для context), `pkg/` (types).
   Не зависит от: `engagement/`, `billing/`, `nats/` (триггеры — Go callback, events — через callback interface).

## Примеры CEL-условий

| Русское условие | CEL expression |
|----|--|
| Заполнил >= 5 опросов | `game.engagement_by_category['survey'] >= 5` |
| Стаж >= 3 года | `user.years_of_service >= 3` |
| Пореферал >= 3 человека | `game.engagement_by_category['referral'] >= 3` |
| Активен сегодня | `game.days_since_active <= 1` |
| Подключил льготы из >= 3 категорий | `game.benefit_categories_count >= 3` |
| Имеет ачивку "Мастер опросов" | `game.achievements.contains('survey_master')` |

> **Важно:** `game.*` использует те же сравнения что и остальные домены. НЕ использовать `.avg()` — функция не зарегистрирована в `cel/functions.go` (доступны только `date_diff_days()`, `str_contains()`, `now_iso()`).

## Триггеры присвоения

| Событие | Источник | Что проверяет | Как часто |
|---------|----------|---------------|-----------|
| `engagement_completed` | Go callback: `FlowEngine.Complete()` → `TriggerHandler.OnEngagementCompleted()` | Все CEL-ачивки с `trigger_on='engagement_completed'` | Каждое completion |
| Monthly cron (Asynq) | Asynq scheduled job | CEL-ачивки с `trigger_on='monthly_cron'` + Loyalty upgrade check | 1-е число месяца |
| Admin API manual-award | HTTP handler | Ручное присвоение (один юзер → одна ачивка) | По запросу HR |
| Admin API batch-import | HTTP handler + Asynq worker | Массовое присвоение из XLSX | По запросу HR |

## Аргументы «за»

- **Единый CEL** — не вводим 5-й механизм условий
- **LLM-генерация** — HR формулирует на русском → CEL автоматически
- **Immutable facts** — audit trail автоматический, нет дубликатов (UNIQUE constraint)
- **Go callback** — проще чем NATS-subscriber для in-process событий
- **XLSX import** — паттерн уже есть в `user/`, легко копировать
- **Обратная связь** — геймификация влияет на eligibility: теги ачивок доступны как `game.achievements`

## Аргументы «против»

- Новый internal пакет (11-й) — оправдан: геймификация не является частью eligibility (другой домен), не часть engagement (не каталог и не flow)
- Расширение CELContext — требует синхронного обновления schema в `cel/context.go` (ADR-021 предписывает синхронный update при изменении CEL context)
- Monthly cron для всех users — может быть дорого при large tenant. Mitigation: проверить только тех, у кого есть pending achievements

## Вердикт

**За.** CEL + immutable facts минимизирует дублирование логики evaluation. Новый пакет `gamification/` чётко отделён SRP. Go-callback триггеры проще NATS-subscriber для in-process событий.

## Следствия

- `cel/context.go` добавить блок `Game struct`
- `cel-engine.md` добавить Gamification в таблицу доменов (строка 5)
- `модули.md` добавить `gamification/` в таблицу пакетов Platform
- `пакеты-platform.md` добавить описание 11-го пакета + DI граф обновление
- `engagement.md` описать Go-callback в FlowEngine.Complete()
- Создать миграции SQL для 5 таблиц
- Создать Asynq worker `gamification-check-monthly`
- Создать Asynq worker `gamification-import-xlsx`

## Статус

✅ Accepted (M09, T0901)
