# T0901 — Архитектура системы геймификации

## Контекст

Платформы гибких льгот не хватает слоя геймификации. Прототип показывает прогресс-бары, badge-статусы, баллы за активности — но backend-механики присвоения ачивок и уровней лояльности не спроектированы.

**Связь с существующей архитектурой:**
- `архитектура/cel-engine.md` — CEL уже есть, используется для evaluation условий
- `архитектура/engagement.md` — UserEngagement.completed → триггер для проверки ачивок
- `контекст/проблема.md` §12 — геймификация как driver вовлечённости

**Ключевое решение:** один из источников присвоения тега/ачивки — CEL-condition evaluation. CEL уже покрывает billing, eligibility, flow, recommendations. Геймификация — 5-й домен, использующий тот же движок.

```
            CEL Engine (условия)              Геймификация (факты)
┌─────────────────────────────────┐      ┌─────────────────────────┐
│ achievements.cel_condition:     │      │ user_achievements:       │
│ "completions >= 5 &&           │──→   │  user_id,                │
│  answers.avg_score >= 7"       │      │  achievement_key,        │
│                                 │      │  awarded_at, visible     │
│ loyalty.cel_condition:         │      └─────────────────────────┘
│ "engagement_count >= 10"       │         ↑
└─────────────────────────────────┘    CheckAchievementEngine
                                         (trigger: engagement events + cron)
```

## Задача

Спроектировать модуль геймификации:

### 1. DB Schema

```sql
-- Ачивки (шаблоны — создаёт HR/админ)
CREATE TABLE achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    key VARCHAR(64) NOT NULL,              -- 'survey_master', 'tenure_3y', 'fitness_freak'
    name VARCHAR(128) NOT NULL,            -- 'Мастер опросов', 'Три года с нами'
    description TEXT,
    icon_url VARCHAR(256),
    category VARCHAR(32),                 -- 'engagement' | 'loyalty' | 'billing' | 'custom'
    tier INTEGER DEFAULT 1,               -- порядок сортировки / редкость
    is_hidden BOOLEAN DEFAULT false,       -- пасхалки — не показываются до получения

    -- Условия присвоения
    condition_type VARCHAR(16) NOT NULL DEFAULT 'cel',  -- 'cel' | 'manual' | 'event' | 'xlsx'
    condition_cel TEXT,                   -- CEL expression (если type='cel')
    condition_source TEXT,               -- русский текст для LLM-генерации
    trigger_on VARCHAR(32),              -- 'engagement_completed' | 'monthly_cron' | 'manual' | 'xlsx_import'

    -- Metadata
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(tenant_id, key)
);

-- Присвоенные ачивки (факты — неизменяемые)
CREATE TABLE achievement_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),  -- multi-tenant изоляция
    user_id UUID NOT NULL REFERENCES users(id),
    achievement_id UUID NOT NULL REFERENCES achievements(id),
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    awarded_by VARCHAR(16) NOT NULL DEFAULT 'system', -- 'system' | 'hr' | 'admin' | 'xlsx'
    metadata JSONB,                            -- { completions: 7, engagement_count: 12 }
    UNIQUE(user_id, achievement_id)
);

-- Уровни лояльности (шаблоны уровней)
CREATE TABLE loyalty_level_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    key VARCHAR(32) NOT NULL,                 -- 'bronze', 'silver', 'gold', 'platinum'
    name VARCHAR(64) NOT NULL,
    icon_url VARCHAR(256),
    level_order INTEGER NOT NULL,             -- порядок: 1=bronze, 2=silver, ...
    description TEXT,
    benefits JSONB,                          -- бонусы уровня: { extra_points: 100, exclusive_offers: true }
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(tenant_id, key)
);

-- Текущий уровень пользователя (историчный — valid_to)
CREATE TABLE user_loyalty_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),  -- multi-tenant изоляция
    user_id UUID NOT NULL REFERENCES users(id),
    level_definition_id UUID NOT NULL REFERENCES loyalty_level_definitions(id),
    level_key VARCHAR(32) NOT NULL,           -- duplicа для быстрых запросов
    entered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    exited_at TIMESTAMPTZ,                   -- NULL = текущий уровень
    exit_reason VARCHAR(32),                -- 'downgraded', 'expired', NULL
    metadata JSONB,                          -- { points_at_entry: 4200, achievements: 7 }

    -- Один активный уровень на момент времени
    CONSTRAINT no_overlap EXCLUDE USING gist (
        user_id WITH =,
        tstzrange(entered_at, COALESCE(exited_at, now()), '[]') WITH &&
    )
);

-- Индексы
CREATE INDEX idx_achievement_grants_user ON achievement_grants(user_id, tenant_id);
CREATE INDEX idx_achievement_grants_achievement ON achievement_grants(achievement_id);
CREATE INDEX idx_user_loyalty_current ON user_loyalty_levels(user_id, tenant_id) WHERE exited_at IS NULL;
CREATE INDEX idx_user_loyalty_history ON user_loyalty_levels(tenant_id, exited_at);
CREATE INDEX idx_import_jobs_started_by ON gamification_import_jobs(started_by);
```

### 2. Модуль `gamification/` в Platform

```
platform/internal/gamification/
├── models.go          # Achievement, AchievementGrant, LoyaltyLevel, UserLoyaltyLevel
├── achievement.go     # AchievementEngine: Create, List, GetUserAchievements, Award
├── grant_engine.go    # GrantEngine: CheckAndAward(ctx, userId) — проверка ВСЕХ условий
├── loyalty.go         # LoyaltyEngine: GetCurrentLevel, Upgrade, Downgrade, CheckEligibility
├── triggers.go        # TriggerHandler: OnEngagementCompleted, OnMonthlyCron, AwardManually
├── api_handlers.go    # HTTP handlers (CRUD achievements, user badges, leaderboard)
└── cel_integration.go # BuildGamificationCELContext() — расширение CELContext
```

### 3. Интеграция с CEL

CELContext расширяется полями для геймификации (вложенный блок `game.*`):

```go
// Добавляем в cel/context.go:
type CELContext struct {
    // ... существующие поля (user, tags, benefit, date, context, answers, events, period, balance) ...

    // Геймификация — вложенный (game.*)
    Game struct {
        Achievements         []string         `cel:"achievements"`           // ключи имеющихся ачивок
        AchievementCount     int              `cel:"achievement_count"`     // количество ачивок
        EngagementCount      int              `cel:"engagement_count"`      // всего завершённых энгейджментов
        EngagementByCategory map[string]int   `cel:"engagement_by_category"` // map[string]int: {'survey': 3, 'referral': 2}
        BenefitCategories    int              `cel:"benefit_categories_count"` // кол-во категорий льгот, в которые юзер подключен
        LoyaltyLevel         string           `cel:"loyalty_level"`         // текущий уровень
        LoyaltyPoints        float64          `cel:"loyalty_points"`        // cumulative engagement points
        DaysSinceActive      int              `cel:"days_since_active"`    // дней с последнего актив
        EnpsSubmitted        bool             `cel:"enps_submitted"`
        HasFamily            bool             `cel:"has_family"`      // есть родственники в системе ДМС
    } `cel:"game"`
}
```

Примеры CEL-условий ачивок:
```cel
// "Мастер опросов": заполнил >= 5 опросов
game.engagement_by_category['survey'] >= 5

// "Три года с нами" (user.* из HR-данных)
user.years_of_service >= 3

// "Социальный": пореферал >= 3 человек
game.engagement_by_category['referral'] >= 3

// "Регулярный пользователь": активен последний день
game.days_since_active <= 1

// "Универсал": подключил льготы из >= 3 категорий
game.benefit_categories_count >= 3
```

> **Важно:** `game.*` использует те же сравнения что и остальные домены. НЕ использовать `.avg()` — функция не зарегистрирована в `cel/functions.go` (доступны только `date_diff_days()`, `str_contains()`).

### 4. Триггеры присвоения

Gamification получает события **внутри Platform** — не подписывается на NATS billing subjects (consumer = billing service, не platform/gamification). Интеграция через Go-callback из `engagement/flow/FlowEngine.Complete()`.

| Событие | Источник | Что проверяет | Как часто |
|---------|--|-----|----|
| `engagement_completed` | Go callback: `FlowEngine.Complete()` → `TriggerHandler.OnEngagementCompleted()` | Все CEL-ачивки с `trigger_on='engagement_completed'` | Каждое completion |
| Monthly cron (Asynq) | Asynq scheduled job | CEL-ачивки с `trigger_on='monthly_cron'` + Loyalty upgrade check | 1-е число месяца |
| Admin API `POST /gamification/v1/manual-award` | HTTP handler | Ручное присвоение (один юзер → одна ачивка) | По запросу HR |
| Admin API `POST /gamification/v1/batch-import` | HTTP handler + Asynq worker | Массовое присвоение из XLSX | По запросу HR |

> **Почему billing.credit не триггер:** billing credit — это финансовый результат engagement completion. Gamification получает то же событие из FlowEngine.Complete() — до того как NATS-событие опубликовано. Это проще: нет отдельного NATS subscriber, нет race-condition "credit начислен но ачивка не присвоена".

### 5. API

Все endpoints используют cursor pagination (spec/api.md: `nextCursor`). RBAC guard — middleware на уровне `api/`.

| Endpoint | Method | Роли | Описание |
|-----|--|--|-|
| `/gamification/v1/achievements` | GET | hr, admin | Список ачивок tenant'а |
| `/gamification/v1/achievements` | POST | hr, admin | Создание ачивки |
| `/gamification/v1/achievements/:id` | GET | hr, admin | Детали |
| `/gamification/v1/achievements/:id` | PUT | hr, admin | Редактирование |
| `/gamification/v1/achievements/:id` | DELETE | hr, admin | Удаление (soft: не удалять гранты) |
| `/gamification/v1/user/achievements` | GET | employee, hr, admin | Мои ачивки (сотр = свой uid) |
| `/gamification/v1/user/progress` | GET | employee, hr, admin | Прогресс к НЕПОЛУЧЕННЫМ ачивкам (каждая = current/needed) |
| `/gamification/v1/user/level` | GET | employee, hr, admin | Текущий уровень лояльности |
| `/gamification/v1/user/level/history` | GET | employee, hr, admin | История уровней |
| `/gamification/v1/manual-award` | POST | hr, admin | Ручное присвоение (body: userId, achievementId) |
| `/gamification/v1/cel/generate-achievement` | POST | hr, admin | LLM генерация CEL-условия (body: sourceText) |

### 6. Массовое присвоение — XLSX Import

HR-менеджер часто получает списки «кому дать ачивку» из других систем (1С, SAP, Excel от руководства). Механика аналогична `user/registry-import-xlsx` (T0101), но адаптирована под геймификацию.

#### Workflow

```
HR загружает XLSX
    │
    ├─ → валидация формата (колонки, типы)
    │    ├─ ошибка формата → скачать отчёт (строка, колонка, причина)
    │    └─ OK
    │
    ├─ → валидация данных (user_id exists, achievement exists, нет дубликатов)
    │    ├─ ошибки → скачать отчёт (строка, user_id, причина)
    │    └─ OK
    │
    ├─ → preview: «X записей валидно, Y с ошибками — применить?»
    │    └─ админ подтверждает
    │
    └─ → массовая вставка (achievement_grants) + NATS event per granted user
```

#### Формат XLSX

Два шаблона (админ скачивает пустой шаблон перед заполнением):

**Шаблон A — Массовое присвоение ачивки:**

| A: email | B: full_name | C: achievement_key | D: comment |
|---|---|---|---|
| ivanov@mail.ru | Иванов И.И. | survey_master | Участник Q3 |
| petrov@mail.ru | Петров П.П. | survey_master | Участник Q3 |

- `email` — корпоративный email сотрудника → поиск по `users.email` (уникальный identifier, см. акторы.md)
- `full_name` — для визуальной проверки (не используется для поиска)
- `achievement_key` — ключ ачивки из `achievements.key`
- `comment` — записывается в `achievement_grants.metadata.comment`

**Шаблон B — Массовое изменение уровней лояльности:**

| A: email | B: full_name | C: new_level_key | D: reason |
|---|---|---|---|
| ivanov@mail.ru | Иванов И.И. | gold | По рекомендации директора |
| petrov@mail.ru | Петров П.П. | silver | Перевод из бронзы |

- `new_level_key` — ключ уровня из `loyalty_level_definitions.key`
- При успехе: текущий уровень закрывается (`exited_at = now`, `exit_reason = 'xlsx_import'`), открывается новый
- `reason` → `user_loyalty_levels.metadata.reason`

#### API для XLSX import

```
POST   /gamification/v1/imports/batch           — загрузить XLSX (async, возвращает import_id)
GET    /gamification/v1/imports/:id/status      — статус импорта (pending/validating/preview/ready/done/failed)
GET    /gamification/v1/imports/:id/errors      — скачать отчёт об ошибках (XLSX)
GET    /gamification/v1/imports/:id/preview     — preview валидных записей (JSON, пагинация)
POST   /gamification/v1/imports/:id/apply       — применить импорт (требует подтверждения)
GET    /gamification/v1/imports/templates/badges   — скачать пустой шаблон A
GET    /gamification/v1/imports/templates/levels   — скачать пустой шаблон B
```

#### Asynq worker

```
`gamification-import-xlsx` → gamification/ImportXLSX
```

Аналогично `registry-import-xlsx` (user/). Этапы: валидация → preview → apply → запись в `import_jobs` → NATS events.

#### Таблица import_jobs (общая с user import? или своя?)

Своя, чтобы не смешивать домены, но с тем же паттерном:

```sql
CREATE TABLE gamification_import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),  -- multi-tenant изоляция
    import_type VARCHAR(16) NOT NULL,  -- 'badges' | 'levels'
    file_name VARCHAR(256) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',  -- pending, validating, preview, ready, applying, done, failed
    total_rows INTEGER DEFAULT 0,
    valid_rows INTEGER DEFAULT 0,
    error_rows INTEGER DEFAULT 0,
    error_report_url VARCHAR(512),   -- ссылка на скачиваемый XLSX с ошибками
    started_by UUID REFERENCES users(id),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    error_message TEXT
);
```

### 7. Интеграция с UI

Frontend получает от API:
- Список бейджей пользователя (иконки, tooltip с date)
- Прогресс-бар к следующей ачивке (`{ current: 3, needed: 5, key: 'survey_master' }`)
- Текущий уровень лояльности + бонусы уровня
- Leaderboard (опционально, top-N)

## Ожидаемый результат

После выполнения T0901:
1. ADR-023 записан в `doc/архитектура/adr/023-gamification-system.md`
2. DB schema определена (achievements, achievement_grants, loyalty_levels, import_jobs)
3. Модуль `gamification/` спроектирован (публичный API, trigger points, CEL-context expansion, XLSX import)
4. Интеграционные точки с engagement, billing, notification определены
5. Документация обновлена: `модули.md`, `пакеты-platform.md`, `cel-engine.md` (CELContext)
