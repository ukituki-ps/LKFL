# ADR-041: Secrets Management Roadmap

| Поле     | Значение |
|----------|----------|
| Status   | Accepted |
| Date     | 2026-06-02 |
| Веха     | DevOps Hardening |
| Авторы   | architect-lkfl |
| Related  | ADR-030, ADR-036 |

## Context

Секреты в проекте LKFL:
- `POSTGRES_PASSWORD` — пароль PostgreSQL
- `REDIS_PASSWORD` — пароль Redis (production)
- `KEYCLOAK_ADMIN_PASSWORD` — пароль администратора Keycloak
- `KEYCLOAK_CLIENT_SECRET` — OAuth2 client secret
- `SENTRY_DSN` — DSN для Sentry
- `GHCR_PAT` — токен для Docker registry
- `DEPLOY_TOKEN` — аутентификация Deploy Worker webhooks
- `WEBHOOK_SECRET` — подпись provider webhooks

**Проблемы текущего подхода (.env files):**
- Нет rotation policy
- Нет audit trail (кто и когда менял секреты)
- Секреты на файловой системе сервера (файловый доступ = доступ к секретам)
- При компрометации сервера — все секреты открыты
- Нет автоматической ротации
- `.env.staging` и `.env.prod` — plain text на сервере

## Decision

### Phase 1: .env files + GH Actions (текущее — до M38)

**Что есть:**
- `.env.staging` / `.env.prod` файлы на сервере (в `.gitignore`)
- `.env.staging.example` / `.env.prod.example` — шаблон в git
- GH Actions secrets для CI/CD pipeline
- Секреты передаются через SSH deploy

**Правила Phase 1:**

| Правило | Описание |
|---------|----------|
| `.env.*` в `.gitignore` | Никогда не коммитить реальные секреты |
| `.env.*.example` в git | Шаблон с placeholder-значениями |
| Rotation при инциденте | Ручная замена всех затронутых секретов |
| Audit trail | Server access log + git log для .example |
| Минимальный доступ | SSH key только для deploy пользователей |
| Никогда в чат | Секреты не передаются в чат/коммиты/логи |

**Rotation procedure Phase 1:**

```bash
# 1. Сгенерировать новый секрет
openssl rand -hex 32

# 2. Обновить .env.staging
#    POSTGRES_PASSWORD=новый_пароль

# 3. Если меняется DB пароль → обновить DB
docker exec lkfl-staging-postgres-1 psql -U lkfl -c "ALTER USER lkfl WITH PASSWORD 'новый_пароль';"

# 4. Обновить GH Actions secrets (если используется в CI)

# 5. Перезапустить сервисы
docker compose -f docker-compose.staging.yml -p lkfl-staging restart
```

### Phase 2: Mozilla SOPS + age (M38 — F3 Hardening)

**Решение:**
- Шифрование `.env` файлов с помощью [Mozilla SOPS](https://github.com/mozilla/sops) + [age](https://github.com/FiloSottile/age)
- Зашифрованные `.env.staging.enc` / `.env.prod.enc` в git
- Decrypt на deploy time (в CI runner и Deploy Worker)
- Key rotation через age public keys

**Архитектура:**

```
Developer → sops encrypt → .env.staging.enc → git commit
                                              ↓
CI Runner → sops decrypt → .env.staging → docker compose up
                                              ↓
Deploy Worker → sops decrypt → .env.staging → docker compose up
```

**Конфигурация:**

```yaml
# .sops.yaml
creation_rules:
  - path_regex: env/staging\.env\.enc$
    age: age1XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  - path_regex: env/prod\.env\.enc$
    age: age1YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY
```

**Преимущества Phase 2:**
- Секреты в git (encrypted) — аудит через git log
- Key rotation — замена age key → re-encrypt
- Нет plain text секретов на сервере (decrypt в memory при deploy)

### Phase 3: HashiCorp Vault (M44 — F4 Hardening, production)

**Решение:**
- Vault server на отдельном контейнере/VM
- Dynamic secrets для PostgreSQL (temporary credentials per session)
- AppRole auth для сервисов (lkfl-server, proxy authenticate to Vault)
- Auto-rotation (TTL-based credentials)
- Audit log в Vault (кто, когда, какой секрет)

**Архитектура:**

```
lkfl-server → Vault Agent → dynamic PG credentials (TTL 1h, auto-renew)
lkfl-proxy  → Vault Agent → Redis credentials
Deploy Worker → Vault Agent → decrypt at deploy time
```

**Dynamic PostgreSQL credentials:**

```go
// Vault dynamic DB credential
func GetDBCredentials(ctx context.Context) (string, error) {
    secret, err := vaultClient.Logical().Write("database/creds/lkfl-app", map[string]interface{}{})
    // Returns: username (temporary), password, lease_id
    // Auto-expire after TTL → credential becomes invalid
}
```

**Преимущества Phase 3:**
- Zero-touch rotation — credentials expire automatically
- Audit trail — кто, когда, какой секрет
- Compromise containment — скомпрометированный ключ истекает за TTL
- Centralized management — один Vault для всех сервисов

### Сравнение фаз

| Критерий | Phase 1 (.env) | Phase 2 (SOPS) | Phase 3 (Vault) |
|----------|---------------|----------------|-----------------|
| Storage | .env на сервере | .enc в git | Vault server |
| Audit | server log | git log | Vault audit |
| Rotation | Manual | Quarterly | Auto (TTL) |
| Compromise impact | High (permanent) | Medium (re-encrypt) | Low (expire + rotate) |
| Complexity | Low | Medium | High |
| Infra overhead | None | None | Vault server + HA |

## Consequences

### Положительные
- **Градационный подход:** каждая фаза — improvement, не big-bang
- **Phase 1 работает сейчас:** нет блокировки разработки
- **Phase 2 баланс:** security vs complexity для staging
- **Phase 3 production-ready:** dynamic secrets, auto-rotation, audit

### Отрицательные
- **Phase 1:** manual overhead, no audit trail quality
- **Phase 2:** need sops+age binary в CI runner и Deploy Worker
- **Phase 3:** отдельный сервер для Vault HA, operational complexity
- **Migration между фазами:** нужно переписать deploy workflow

### Риски

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| Потеря age key (Phase 2) | Низкая | Multiple key holders + recovery procedure |
| Vault downtime (Phase 3) | Средняя | Vault HA + sealed state monitoring |
| Developer accidentally commits .env | Средняя | pre-commit hook + CI check |
| Секрет в логах/чате | Средняя | Секция «Секреты» в AGENTS.md + lint |

## Related ADR

- **ADR-030:** CI/CD Pipeline — GH Actions secrets для CI
- **ADR-036:** Deploy Worker — webhook auth с DEPLOY_TOKEN
