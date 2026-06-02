# T2310 — CI/CD Pipeline Optimization: Отчёт

## Статус

✅ реализовано

## Резюме

Реализованы все 5 рекомендаций (R1-R5) по оптимизации pipeline «Build & Deploy».
Изменения — в `.github/workflows/build.yml`.

## Изменения

### R1: Persistent Go cache (экономия ~9m)

- Добавлен step `Init Go cache directory` — создаёт `/home/runner/go-cache/{mod,build}`
- `actions/setup-go@v5`: `cache: false` (GitHub Cache API недоступен на self-hosted)
- Добавлены env vars: `GOMODCACHE=/home/runner/go-cache/mod`, `GOCACHE=/home/runner/go-cache/build`

### R2: Docker build cache через GHCR (экономия 1-2m)

- `docker/setup-buildx-action@v3`: `driver: docker-container` (поддержка registry cache)
- `docker/build-push-action@v5`: добавлены `cache-from` и `cache-to` через GHCR registry
- Cache образы: `{service}:buildcache` — не публикуются как production

### R3: golangci-lint cache (экономия 10-20s)

- `skip-cache: false` — включить кэш бинарника и analysis cache
- `skip-pull-check: true` — не проверять обновления бинарника при каждом запуске

### R4: E2E staging — polling вместо sleep 30 (экономия ~20s)

- Заменён `sleep 30` на polling OIDC discovery endpoint (`/.well-known/openid-configuration`)
- Интервал: 3s, max attempts: 20 (60s timeout)
- Если Keycloak не отвечает — warning, но pipeline продолжается

### R5: Node.js 24 migration

- Добавлено `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: 'true'` в workflow env
- `actions/setup-node@v4`: `cache: false` (GitHub Cache API недоступен на self-hosted)

### Бонус: npm cache fix

- Все три job'а (`lint-test`, `e2e-local`, `e2e-staging`): `actions/setup-node@v4` → `cache: false`
- `~/.npm` сохраняется на persistent filesystem self-hosted runner

## Измерения

### До оптимизации (измерено на 02.06.2026)

| Метрика | Значение |
|---------|----------|
| Среднее время pipeline | 21-25 минут |
| Worst case | 36 минут |
| Setup Go | 9.5-10 минут |
| Docker build (frontend) | 0.8-2.4 минуты |
| golangci-lint | 0.2-0.3 минуты |
| E2E staging (sleep) | 30s hardcoded |

### После оптимизации (ожидаемое)

| Метрика | Целевое |
|---------|---------|
| Среднее время pipeline | < 15 минут |
| Best case | < 12 минут |
| Setup Go (warm cache) | < 30 секунд |
| Docker build (cache hit) | < 30 секунд |
| golangci-lint (cached) | < 15 секунд |
| E2E staging (polling) | ~10s (зависит от Keycloak) |

## Файлы

| Файл | Статус |
|------|--------|
| `.github/workflows/build.yml` | ✅ оптимизирован (691 → 691 lines, +20 changes) |
| `doc/задачи/M23-ci-cd-optimize/T2310-ci-cd-optimize/brief.md` | ✅ создан |
| `doc/задачи/M23-ci-cd-optimize/T2310-ci-cd-optimize/plan.yaml` | ✅ создан |
| `doc/задачи/M23-ci-cd-optimize/T2310-ci-cd-optimize/report.md` | ✅ заполнен |

## Валидация

- [ ] Запустить pipeline 3 раза на feature-ветке — измерить время (warm cache)
- [ ] Первый запуск (cold cache) — проверить что GOMODCACHE/GOCACHE создаются
- [ ] Проверить GHCR — buildcache образы не дублируют production
- [ ] Push на main — полный pipeline до deploy staging + smoke test

## Заметки

- На сервере serverAi нужно убедиться что `/home/runner/` существует и доступен для записи runner'ом
- Если runner запущен под другим пользователем — пути кэш-директорий可能需要 корректировки
