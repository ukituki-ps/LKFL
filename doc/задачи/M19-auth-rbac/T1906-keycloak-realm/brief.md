# T1906 — Keycloak Realm Config

## Веха

M19-auth-rbac

## Тип

code

## Контекст

Keycloak realm configuration для tenant «sdek».
Создаётся через docker volume mount realm JSON.

**Текущее состояние docker-compose.yml (M17):**
- Keycloak 25.0, image `quay.io/keycloak/keycloak:25.0`
- DB: PostgreSQL (`KC_DB: postgres`, schema `keycloak` в том же PG)
- Volume: `lkfl_keycloak_data:/opt/keycloak/data`
- Command: `start-dev --http-relative-path=/ --hostname=localhost --hostname-port=8081`
- Порт: 8081
- Issuer URL (для lkfl-server): `http://localhost:8081/realms/lkfl-sdek`

**Что нужно изменить в docker-compose.yml:**
- Добавить volume mount для realm JSON: `./infra/keycloak/realm-lkfl-sdek.json:/opt/keycloak/data/import/realm-lkfl-sdek.json:ro`
- Добавить `--import-realm` в command Keycloak
- Realm создаётся при первом запуске контейнера

## Что сделать

### `infra/keycloak/realm-lkfl-sdek.json`

```json
{
  "id": "lkfl-sdek",
  "realm": "lkfl-sdek",
  "enabled": true,
  "sslRequired": "external",
  "loginTheme": "keycloak",
  "accountTheme": "keycloak",

  "clients": [
    {
      "clientId": "lkfl-spa",
      "enabled": true,
      "publicClient": true,
      "redirectUris": [
        "http://localhost:5173/*",
        "http://localhost:8080/*",
        "http://lkfl-staging.example.com/*"
      ],
      "webOrigins": [
        "http://localhost:5173",
        "http://localhost:8080"
      ],
      "protocol": "openid-connect",
      "standardFlowEnabled": true,
      "implicitFlowEnabled": false,
      "directAccessGrantsEnabled": true,
      "serviceAccountsEnabled": false,
      "attributes": {
        "pkce.code.challenge.method": "S256"
      }
    },
    {
      "clientId": "lkfl-service",
      "enabled": true,
      "publicClient": false,
      "clientAuthenticatorType": "client-secret",
      "secret": "${KEYCLOAK_SERVICE_SECRET}",
      "serviceAccountsEnabled": true,
      "protocol": "openid-connect",
      "attributes": {
        "client.secret.creation.time": "1700000000"
      }
    }
  ],

  "roles": {
    "realm": [
      {
        "id": "role-employee",
        "name": "employee",
        "description": "Сотрудник"
      },
      {
        "id": "role-hr",
        "name": "hr",
        "description": "HR-менеджер"
      },
      {
        "id": "role-catalog-manager",
        "name": "catalog_manager",
        "description": "Менеджер каталога"
      },
      {
        "id": "role-admin",
        "name": "admin",
        "description": "Администратор платформы"
      }
    ]
  },

  "users": [
    {
      "id": "admin-user",
      "username": "admin",
      "enabled": true,
      "email": "admin@lkfl.dev",
      "credentials": [
        {
          "type": "password",
          "value": "admin-dev-password",
          "temporary": false
        }
      ],
      "realmRoles": ["admin"]
    }
  ]
}
```

### Docker integration

В `docker-compose.yml` изменить секцию `keycloak`:

```yaml
keycloak:
  command: >
    start-dev --import-realm
    --http-relative-path=/
    --hostname=localhost
    --hostname-port=8081
    --hostname-strict=false
  volumes:
    - lkfl_keycloak_data:/opt/keycloak/data
    - ./infra/keycloak/realm-lkfl-sdek.json:/opt/keycloak/data/import/realm-lkfl-sdek.json:ro
```

> **Важно:** `--import-realm` обрабатывает все `.json` файлы в `/opt/keycloak/data/import/`.
> Volume `lkfl_keycloak_data` уже смонтирован на `/opt/keycloak/data` (M17).
> Realm JSON монтируется как read-only (`:ro`) в поддиректорию `import/`.

### Keycloak Admin API script (альтернатива для production)

### Keycloak Admin API script (альтернатива)

```bash
# infra/keycloak/setup-realm.sh
#!/bin/bash
# Создаёт realm через Admin API (для production, где import-realm недоступен)
curl -X POST "$KEYCLOAK_ADMIN_URL/admin/realms" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d @infra/keycloak/realm-lkfl-sdek.json
```

## Требования

- Realm: `lkfl-sdek`
- Client: `lkfl-spa` (public, PKCE S256)
- Client: `lkfl-service` (confidential, service account)
- Roles: employee, hr, catalog_manager, admin
- Dev user: admin (temporary password)
- Redirect URIs: localhost dev + staging domain
- PKCE S256 required

## Критерии приёмки

- [ ] `infra/keycloak/realm-lkfl-sdek.json` создан
- [ ] Realm импортируется через docker volume
- [ ] Client `lkfl-spa` (public, PKCE)
- [ ] Client `lkfl-service` (confidential)
- [ ] Roles: employee, hr, catalog_manager, admin
- [ ] Dev admin user
- [ ] Login через Keycloak работает
