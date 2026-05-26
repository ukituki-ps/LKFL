# Отчёт — T0901

## Выполнено

1. **ADR-023 создан** — `doc/архитектура/adr/023-gamification-system.md`
   - Контекст: платформы не хватает слоя геймификации
   - Два варианта рассмотрены: Tag Engine (отдельный) vs CEL + immutable facts (выбрано)
   - Ключевые решения: CEL 5-й домен, immutable grants, Go-callback триггеры, XLSX import, отдельный пакет

2. **DB schema определена** — 5 таблиц в ADR-023 §Schema:
   - `achievements` — шаблоны ачивок (key, name, cel_condition, trigger_on)
   - `achievement_grants` —Immutable факты присвоения (UNIQUE user + achievement)
   - `loyalty_level_definitions` — шаблоны уровней (bronze/silver/gold/platinum)
   - `user_loyalty_levels` — историчные уровни пользователя (valid_to pattern, exclusion constraint)
   - `gamification_import_jobs` — трекинг XLSX импортов (status, rows, errors)

3. **Модуль `gamification/` спроектирован:**
   - 7 файлов: models.go, achievement.go, grant_engine.go, loyalty.go, triggers.go, cel_integration.go, xlsx_import.go
   - Public API: 12 методов (CRUD achievements, user-facing API, grant engine, loyalty, admin)
   - Trigger contract: EngagementCompleter interface (Go-callback, best-effort)

4. **CEL-интеграция описана:**
   - CELContext расширен блоком `game.*` (10 полей)
   - 5-й домен CEL: conditions хранятся в `achievements.condition_cel`
   - Примеры CEL-условий: 5 реальных сценариев на русском с CEL translation

5. **Триггеры описаны:**
   - 4 типа: engagement_completed (Go-callback), monthly_cron (Asynq), admin manual (HTTP), xlsx import (Asynq + HTTP)
   - FlowEngine.Complete() → TriggerHandler.OnEngagementCompleted() — best-effort, не блокирует completion

6. **XLSX import workflow описан:**
   - 2 шаблона: badges (email + achievement_key) + levels (email + new_level_key)
   - Workflow: validate → preview → apply → import_jobs record
   - Asynq worker: `gamification-import-xlsx`

7. **Документация обновлена:**
   - `модули.md` — gamification/ в таблице пакетов, DI граф + callback → gamification/, Asynq workers +2, PostgreSQL + achievements
   - `пакеты-platform.md` — M09 section, 11-й пакет, новый раздел gamification/, HandlerDeps + Gamification, dependencies table updated, workers +2, summary table +gamification
   - `cel-engine.md` — game.* в CELContext, Gamification в таблице доменов (5-й), таблица полей расширена, 14 точек интеграции, system prompt + game schema, 3 новые metrics
   - `engagement.md` — раздел "Интеграция с Геймификацией": Go-callback contract, best-effort semantics

## Проблемы

Нет архитектурных проблем.

## Следующие шаги

T0901 завершён. М09-gamification-system — архитектура спроектирована.
Для реализации Go-кода требуется отдельная задача (M10 или дальнейшие вехи).
