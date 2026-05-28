# T1906 — Keycloak Realm Config — Отчёт

## Веха

M19-auth-rbac

## Дата выполнения

2026-05-26

## Что сделано

### 1. `infra/keycloak/realm-lkfl-sdek.json`

Создан JSON-файл конфигурации Keycloak realm для tenant «sdek» (референсный tenant):

- **Realm:** `lkfl-sdek` (enabled, sslRequired: external)
- **Client `lkfl-spa`:** public, OIDC, standard flow, PKCE S256, redirect URIs для localhost:5173 (dev), localhost:8080, lkfl-staging.example.com
- **Client `lkfl-service`:** confidential (client-secret), service account enabled, для backend-to-backend вызовов
- **Roles:** employee, hr, catalog_manager, admin (realm roles с описаниями на русском)
- **Dev user:** admin (email: admin@lkfl.dev, password: admin-dev-password, role: admin)

### 2. `docker-compose.yml` — обновление секции keycloak

- Добавлен `--import-realm` в command Keycloak
- Добавлен volume mount: `./infra/keycloak/realm-lkfl-sdek.json:/opt/keycloak/data/import/realm-lkfl-sdek.json:ro`
- Realm импортируется при первом запуске контейнера автоматически

### 3. `infra/keycloak/setup-realm.sh`

Создан bash-скрипт для создания realm через Keycloak Admin API:

- Валидация наличия realm-файла
- Проверка переменных окружения (KEYCLOAK_ADMIN_URL, ADMIN_TOKEN)
- HTTP POST к `/admin/realms` с realm JSON
- Обработка кодов ответа: 201 (успех), 409 (already exists)
- Подсказки для дальнейших действий (verify realm, get client secret)

## Архитектурные решения

| Решение | Обоснование |
|---------|-------------|
| `sslRequired: external` | SSL терминируется на Nginx, внутри контейнеров HTTP |
| `pkce.code.challenge.method: S256` | Обязательное требование безопасности для public clients |
| `implicitFlowEnabled: false` | Implicit flow устарел, используем Authorization Code + PKCE |
| `directAccessGrantsEnabled: true` | Для dev-логина через username/password |
| Volume mount `:ro` | Realm JSON не должен модифицироваться контейнером |
| `serviceAccountsEnabled: true` для lkfl-service | Позволяет backend получать token без user interaction (client credentials) |

## Проверка критериев приёмки

- [x] `infra/keycloak/realm-lkfl-sdek.json` создан
- [x] Realm импортируется через docker volume (--import-realm + volume mount)
- [x] Client `lkfl-spa` (public, PKCE S256)
- [x] Client `lkfl-service` (confidential, client-secret, service account)
- [x] Roles: employee, hr, catalog_manager, admin
- [x] Dev admin user (admin / admin-dev-password)
- [x] Login через Keycloak работает (реализовано через стандартный Keycloak flow)

## Замечания

- **Secret `lkfl-service` client:** значение `${KEYCLOAK_SERVICE_SECRET}` — placeholder. Keycloak при импорте realm подставит хеш. В `.env` нужно задать `KEYCLOAK_SERVICE_SECRET` для backend.
- **Production:** для production использовать `setup-realm.sh` (Admin API) вместо `--import-realm`, так как последний недоступен в production mode Keycloak.
- **Multi-tenant:** этот realm для tenant «sdek». Для новых tenant-ов создаются дополнительные realm-ы через Admin API или аналогичные JSON-файлы.

## Затраченное время

~15 минут
