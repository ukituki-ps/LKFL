# T1714 — Staging CI/CD: миграции, realm JSON, schema fix

## Веха

M17-authorization

## Тип

code

## Контекст

Во время деплоя T1711 (user provisioning) на staging были выявлены системные проблемы CI/CD:

1. **Миграции не применяются автоматически** — при push на main CI/CD собирает образ и перезапускает контейнеры, но не выполняет `lkfl-server migrate`. На staging DB осталась старая схема (`keycloak_user_id` вместо `keycloak_sub`, missing columns в `user_roles`).

2. **Realm JSON не обновляется** — `infra/keycloak/realm-lkfl-sdek.json` копируется в container volume при первом создании контейнера. При push нового JSON в git он не попадает на staging. При перезапуске Keycloak импортирует старый realm: `Realm 'lkfl-sdek' already exists. Import skipped`.

3. **Keycloak hostname URL не настроен** — issuer в JWT токене содержит внутренний порт (`http://dev.april.ukituki.tech:19081/realms/lkfl-sdek`) вместо публичного URL. Сервер не может сделать OIDC discovery через публичный URL.

4. **Schema mismatch в user_roles** — код ожидает колонки `granted_at`, `granted_by`, `expires_at`, но migration не создала их.

## Что делать

### 1. Добавить автоматическое применение миграций в CI/CD

**Файл:** `.github/workflows/build.yml`

Добавить job `run-migrations` между `deploy-staging` и `smoke-test-staging`:

```yaml
run-migrations:
  needs: [deploy-staging]
  runs-on: self-hosted
  steps:
    - name: Apply migrations on staging
      run: |
        ssh ukituki@192.168.1.46 "
          cd /home/ukituki/LKFL-staging &&
          docker compose -f docker-compose.staging.yml --env-file .env.staging exec lkfl-server lkfl-server migrate
        "
```

**Обоснование:** Без этого код и схема БД рассинхронизируются. Каждая миграция должна применяться автоматически при деплое.

### 2. Обновлять realm JSON при деплое

**Файл:** `.github/workflows/build.yml`

Добавить шаг в `deploy-staging` до перезапуска Keycloak:

```yaml
- name: Update Keycloak realm config
  run: |
    scp infra/keycloak/realm-lkfl-sdek.json ukituki@192.168.1.46:/home/ukituki/LKFL-staging/infra/keycloak/
```

Или (лучше) добавить volume mount для realm JSON в compose:

```yaml
keycloak:
  volumes:
    - ./infra/keycloak/realm-lkfl-sdek.json:/opt/keycloak/data/import/realm-lkfl-sdek.json:ro
```

**Обоснование:** Realm JSON — часть конфигурации, должен обновляться вместе с кодом.

### 3. Fix Keycloak hostname URL

**Файл:** `docker-compose.staging.yml`

Добавить в Keycloak service:

```yaml
KC_HOSTNAME_URL: https://dev.april.ukituki.tech
```

**Обоснование:** Без этого issuer в JWT содержит внутренний порт → OIDC discovery падает → сервер не может верифицировать токены.

### 4. Добавить migration для user_roles columns

**Файл:** `migrations/20260531090001_user_roles_columns.sql`

```sql
-- Ref: T1714 — Add missing columns to user_roles
ALTER TABLE lkfl_platform.user_roles
    ADD COLUMN IF NOT EXISTS granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS granted_by UUID REFERENCES lkfl_platform.users(id),
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
```

**Обоснование:** Код `AddRole` ожидает эти колонки. Без migration CI/CD упадёт на staging.

### 5. Добавить schema health check в CI/CD

**Файл:** `.github/workflows/build.yml`

Добавить step после migrations:

```yaml
- name: Verify schema matches code
  run: |
    ssh ukituki@192.168.1.46 "
      docker exec lkfl-staging-postgres-1 psql -U lkfl -d lkfl_platform -c '
        SELECT column_name FROM information_schema.columns
        WHERE table_name = '\''users'\'' AND column_name = '\''keycloak_sub'\''
      '"
```

**Обоснование:** Раннее обнаружение рассинхронизации схемы.

## Требования

- Миграции применяются автоматически при каждом push на main
- Realm JSON обновляется на staging при каждом push
- Keycloak issuer соответствует публичному URL
- Все колонки в БД соответствуют коду
- Schema health check в CI/CD

## Критерии приёмки

- [ ] Job `run-migrations` в CI/CD применяет все миграции
- [ ] Realm JSON обновляется на staging при push
- [ ] Keycloak issuer = `https://dev.april.ukituki.tech/realms/lkfl-sdek`
- [ ] Migration `20260531090001_user_roles_columns.sql` применена
- [ ] Schema health check в CI/CD
- [ ] E2E: login → /me → 200 OK на staging после push

## Зависимости

- **depends_on:** T1711 (user provisioning)
- **touches:** `.github/workflows/build.yml`, `docker-compose.staging.yml`, `migrations/`
- **risk:** Изменение CI/CD пайплайна — может сломать деплой если migration упадёт
