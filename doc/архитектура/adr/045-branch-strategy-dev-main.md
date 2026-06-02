# ADR-045: Двухветочная стратегия — dev → staging, main → production

| Поле     | Значение |
|----------|---------|
| Статус   | Accepted |
| Дата     | 2026-06-02 |
| Авторы   | devops-lkfl |

## Контекст и проблема

Проект использовал стратегию с feature-ветками: разработчики создавали `feature/*` ветки,
merge-ли их в `main`, а `main` автоматически деплоился на staging.

**Проблемы:**
1. **Потеря функционала** — несколько раз функционал терялся в старых feature-ветках,
   которые не были своевременно слиты в main.
2. **Накопление устаревших веток** — feature-ветки отставали от main, merge конфликты росли.
3. **Нет быстрого feedback** — разработчик работает в feature-ветке, но staging обновляется
   только после merge → долго видно результат.
4. **Избыточная сложность** — управление множеством веток при малой команде (1-2 разработчика).

## Решение

Оставить ровно **две ветки**: `dev` и `main`. Все работают в `dev`, она всегда идёт на staging.
`main` — стабильная, используется для production.

### Стратегия

| Ветка | Деплой | CI/CD | Назначение |
|-------|--------|-------|-----------|
| **dev** | staging (serverAi / project.ukituki.tech) | Auto: build → deploy-staging → smoke-test → e2e | Основная рабочая ветка, все мержат сюда напрямую |
| **main** | production (dev.april.ukituki.tech) | Manual: workflow_dispatch | Стабильная версия, мержим из dev когда готово к релизу |

### Workflow

```
Разработчик работает на dev:
  git checkout dev
  git pull origin dev
  # ... работа, коммиты ...
  git push origin dev          → CI/CD → staging автоматически

Когда staging стабилен и готов к production:
  git checkout main
  git merge dev                → CI build+test (но не deploy)
  git push origin main         → GitHub Actions UI → manual production deploy

При проблемах на staging:
  # Исправляешь прямо в dev, пушишь — автоматически обновляется staging

При критических проблемах на production:
  git checkout main
  git merge dev --no-ff        → fix доступен и на staging (через следующий push dev)
  git push origin main         → manual deploy на production
```

### Tag strategy Docker-образов

| Ветка | Image tag | Latest alias |
|-------|-----------|-------------|
| `dev` | `dev-{short-sha}` | `dev-latest` |
| `main` | `main-{short-sha}` | `main-latest` |
| PR | `pr-{number}-{short-sha}` | — |

### CI/CD изменения (build.yml)

**Было:**
- Push в `main` → auto deploy staging
- Push в `feature/*` → только build+test, нет deploy

**Стало:**
- Push в `dev` → auto deploy staging (build-push → deploy-staging → smoke-test → e2e)
- Push в `main` → build+push образов, но без авто-деплоя (production — manual dispatch)
- PR → только CI (lint-test + e2e-local), без push в GHCR

### Branch Protection (рекомендуемые правила GitHub)

| Ветка | Правило | Обоснование |
|-------|---------|------------|
| `dev` | Нет ограничений | Быстрая работа, команда доверяет друг другу |
| `main` | Require PR review (optional) | Защита production от случайных push |
| `main` | Require status checks pass | CI должен быть зелёным перед merge в main |

## Последствия

### Положительные

1. **Нет потери функционала** — все коммиты сразу в dev, ничего не теряется в старых ветках.
2. **Быстрый feedback** — push → staging за ~15-25 минут, результат виден сразу.
3. **Простота** — две ветки вместо множества feature-веток, меньше merge конфликтов.
4. **Staging всегда актуален** — отражает текущее состояние разработки.
5. **Production отделён** — main не меняется без явного решения (merge из dev).

### Отрицательные

1. **Staging может быть нестабилен** — если запушить сломанный код в dev, staging упадёт.
   - Mitigation: CI lint-test должен пройти перед merge. При падении — быстрый rollback или fix push.
2. **Нет изоляции фич на staging** — нельзя тестировать одну фичу отдельно от другой.
   - Mitigation: ручное переключение IMAGE_TAG в `.env.staging` для конкретного образа.
3. **Git history в dev линейный** — без feature веток история проще, но меньше контекста "что и зачем".
   - Mitigation: качественные commit messages + связка с задачами T{MM}{NN}.

### Риски

| Риск | Вероятность | Влияние | Митигация |
|------|------------|---------|-----------|
| Слом staging незапланированно | Средняя | Низкое (staging — не production) | Быстрый fix push или rollback |
| Забытый merge dev → main | Средняя | Среднее (production отстаёт) | Чек-лист релиза, регулярные sync |
| Конфликт при параллельной работе в dev | Низкая (1-2 разработчика) | Низкое | `git pull --rebase` перед push |

## Миграция

### Шаги перехода

1. **Создать ветку dev из main:**
   ```bash
   git checkout -b dev
   git push origin dev
   ```

2. **Обновить CI/CD (build.yml):**
   - Триггеры: `[main, dev]` вместо `[main, 'feature/*']`
   - deploy-staging: `if: github.ref == 'refs/heads/dev'`
   - Tag strategy: `dev-{sha}` + `dev-latest`

3. **Обновить документацию:**
   - `doc/архитектура/deploy-operations.md` — триггеры деплоя
   - Этот ADR

4. **Очистка старых feature-веток (после миграции):**
   ```bash
   # Локально
   git branch | grep 'feature/' | xargs git branch -D

   # На remote (осторожно!)
   git branch -r | grep 'origin/feature/' | sed 's/origin\///' | \
     xargs -I {} git push origin --delete {}
   ```

### Обратимость

Решение полностью обратимо. Если потребуется вернуться к feature-веткам:
1. Откатить build.yml изменения (триггеры обратно на `[main, 'feature/*']`)
2. Разработчики начинают создавать `feature/*` из `dev`
3. ADR помечается как Superseded

## Связанные ADR

- [ADR-030](./030-ci-cd-pipeline.md) — CI/CD Pipeline (базовая архитектура)
- [ADR-039](./039-deploy-rollback-strategy.md) — Rollback Strategy
- [ADR-042](./042-zero-downtime-deployment.md) — Zero-Downtime Deployment
- [ADR-043](./043-two-environment-strategy.md) — Two-Environment Strategy (dev + staging стенды на serverAi)
