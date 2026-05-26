# T1001 — Отложенный recommendations/ → stub — отчёт

## Статус

✅ завершена

## Что сделано

1. **`doc/архитектура/пакеты-platform.md`** — секция `internal/recommendations/` полностью переписана: 8-файльный пакет (engine.go, rules.go, evaluation.go, debug.go, storage.go + full API + Redis cache + PostgreSQL persistence) → 1-файл stub (engine.go: Recommend → empty slice, Debug → nil).
2. **`doc/архитектура/модули.md`** — таблица Platform internal packages: recommendations помечен как stub (Phase 2). Рекомендации удалены из DI графа, зависимость от `cel/`, `user/`, `db/`, Redis убрана.
3. **`doc/архитектура/README.md`** — ссылки на recommendations обновлены: "(stub → Phase 2)".
4. **`doc/контекст/настраиваемость.md`** — «Рекомендации (правила)»: статус изменён на ⚠️ Phase 2.

## Что НЕ трогать

- Файлы Go-кода (backend/internal/recommendations/*.go) — реализация будет добавлена в Phase 2
- DB schema — таблицы recommendation_rules сохранены без изменений

## Проблемы

Нет — задача чистая документация.
