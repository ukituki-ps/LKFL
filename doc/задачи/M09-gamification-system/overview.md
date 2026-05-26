# M09 — Система геймификации

## Описание

Проектирование модуля геймификации: ачивки (achievements), уровни лояльности (loyalty levels), бейджи.

**Ключевой принцип:** CEL — источник условий присвоения. Факты (полученные ачивки, текущий уровень) — неизменяемые записи в БД.

```
CEL Engine (условия)               Геймификация (факты)
├── "стаж >= 3 года"  ──┐          ├── achievement_grants
├── "опросов >= 5"     ├──→       ├── user_loyalty_levels
├── "баллы >= 5000"    ┘          └── gamification_import_jobs
```

### Почему отдельная веха

Геймификация — другой домен:
- **CEL** — evaluation условий в момент запроса (stateless)
- **Геймификация** — факт присвоения (immutable record) + UI-бейджи + прогресс-бары
- Теги (badge-ы) используются **в обратном направлении**: геймификация может влиять на eligibility (`user.achievements.contains('survey_master')`), но не наоборот

### Что будет спроектировано

1. ADR-023: Архитектура системы геймификации
2. DB schema: `achievements`, `achievement_grants`, `loyalty_levels`, `gamification_import_jobs`
3. API: `/gamification/v1/achievements`, `/gamification/v1/levels`, `/gamification/v1/progress`
4. Интеграция с CEL: условия присвоения как CEL-expressions
5. Массовое присвоение: XLSX import (два шаблона — бейджи + уровни), валидация, preview, отчёт об ошибках
6. Интеграция с frontend: бейджи в профиле, прогресс-бары
7. Триггерный движок: когда проверять условия (event-driven через engagement events + cron + manual + xlsx)

## Задачи вехи

| Задача | Описание | Статус |
|-------|--|--|
| T0901 | Архитектура геймификации: ADR-023, DB schema, API, CEL интеграция, XLSX import | ✅ выполнено |
