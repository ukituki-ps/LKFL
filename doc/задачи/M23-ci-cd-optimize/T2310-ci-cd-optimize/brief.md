# T2310 — Оптимизация CI/CD pipeline: Build & Deploy

## Веха

M23 — CI/CD Optimization

## Тип

infra

## Проблема

Pipeline «Build & Deploy» (`build.yml`) на self-hosted runner (serverAi) стабильно занимает **20-25 минут** на успешных запусках, до **36 минут** на проблемных. Это критически замедляет feedback loop при разработке — каждый push на main ждёт 20+ минут до deploy на staging.

### Данные по последним запускам (02.06.2026)

| Запуск | Длительность | Статус |
|--------|-------------|--------|
| 26792143230 | 36m08s | ❌ failure |
| 26811901815 | 25m16s | ✅ success |
| 26798800463 | 21m02s | ✅ success |
| 26797717875 | 22m41s | ✅ success |
| 26791274531 | 20m49s | ✅ success |

### Критический путь

```
Lint & Test (11-14m) → Build-Push (max 2.8m) → Deploy Staging (1.2m)
→ Smoke Test (0.1m) → E2E Staging (6-7m)
= 20-25 минут
```

### Узкие места

**1. 🔴 Setup Go — 9.5-10 минут (45-50% от pipeline)**

`actions/setup-go@v5` с `cache: true` пытается использовать GitHub Cache API, который **не работает на self-hosted runner**. В логах:

```
Failed to restore: Server failed to authenticate the request
Failed to restore: "/usr/bin/tar" failed with error: exit code 2
```

Каждый запуск скачивает Go 1.24 (~100MB) + все модули заново.

**2. 🟠 Docker build без layer cache — 0.5-2.4m**

`docker/build-push-action@v5` вызывается без `cache-from`/`cache-to`. При `driver: docker` (не docker-container) и без explicit cache, каждый build rebuild-ит все слои. Фронтенд при холодном кэше — 2.4m вместо 0.8m.

**3. 🟡 golangci-lint с `skip-cache: true` — ~10s**

Бинарник golangci-lint скачивается при каждом запуске.

**4. 🟡 E2E Staging — `sleep 30` для Keycloak**

Hardcoded sleep 30 секунд вместо polling OIDC discovery endpoint.

**5. 🟢 Node.js 20 deprecated**

Все action'ы показывают warning. GitHub принудительно перейдёт на Node 24 с 16 июня 2026.

## Что делать

### R1: Persistent Go cache на self-hosted runner (экономия ~9m)

Создать persistent volume для Go module cache, отключить GitHub Cache API:

- Настроить `GOMODCACHE` и `GOCACHE` на persistent volume (`/home/runner/go-cache/`)
- Отключить `cache: true` в `actions/setup-go@v5`
- Добавить step для инициализации кэша

### R2: Docker Build Cache через GHCR registry (экономия 1-2m)

Добавить `cache-from`/`cache-to` в `docker/build-push-action@v5`:

```yaml
cache-from: type=registry,ref=${{ env.GHCR_REGISTRY }}/${{ matrix.service }}:buildcache
cache-to: type=registry,ref=${{ env.GHCR_REGISTRY }}/${{ matrix.service }}:buildcache,mode=max
```

### R3: Включить golangci-lint cache

`skip-cache: true` → `skip-cache: false` + `skip-pull-check: true`

### R4: Заменить `sleep 30` на polling в E2E Staging

Заменить hardcoded sleep на polling OIDC discovery endpoint (`/.well-known/openid-configuration`) с интервалом 3s.

### R5: Node.js 24 migration

Добавить `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true` в env workflow или обновить action versions.

## Требования

- Pipeline на успешных запусках — **< 15 минут** (сейчас 20-25m)
- Pipeline на лучших запусках — **< 12 минут** (сейчас ~20m)
- Все изменения — в `.github/workflows/build.yml`
- Backward compatibility — pipeline не должен ломаться при отсутствии cache (первый запуск после очистки)
- Не менять логику deploy, migrations, health check

## Критерии приёмки

- [ ] R1: Go cache persistent volume настроен, `GOMODCACHE`/`GOCACHE` указаны
- [ ] R1: `actions/setup-go@v5` с `cache: false`
- [ ] R2: Docker build с `cache-from`/`cache-to` через GHCR registry
- [ ] R2: Проверить что `driver: docker-container` работает на self-hosted runner
- [ ] R3: `golangci-lint-action` с `skip-cache: false`
- [ ] R4: E2E staging — polling вместо `sleep 30`
- [ ] R5: `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true` добавлено
- [ ] Pipeline работает end-to-end (push на main → deploy staging → smoke test)
- [ ] Измеренное время pipeline (3 запуска) — < 15m среднее
- [ ] Docker images в GHCR не дублируются (buildcache tag не публикуется как production)

## Зависимости

- **depends_on:** нет (самостоятельная задача)
- **touches:** `.github/workflows/build.yml`
- **serverAi:** создание `/home/runner/go-cache/` directory (chmod)
