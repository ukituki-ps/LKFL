# ADR-042: Zero-Downtime Deployment Strategy

| Поле     | Значение |
|----------|----------|
| Status   | Accepted |
| Date     | 2026-06-02 |
| Веха     | DevOps Hardening |
| Авторы   | architect-lkfl |
| Related  | ADR-030, ADR-036, ADR-039 |

## Context

Текущий деплой LKFL (staging на serverAi):
```
docker compose down → pull → migrate → seed → up → health check
```

**Downtime:** ~3-5 минут (compose down + up + health check polling).

**Проблема:** для production с 100K+ сотрудников 5 минут downtime неприемлемо. Docker Compose не поддерживает rolling update нативно. Migrations могут быть breaking change.

## Decision

### Стратегия: Blue-Green via Port Switching

```
┌──────────────────────────────────────────────────────────┐
│                    Nginx (port 80/443)                    │
│                                                           │
│  /api/* → upstream lkfl-server (port 8080)               │
│  / → lkfl-frontend (port 8086)                           │
│  /realms/* → keycloak (port 19081)                        │
└──────────────┬───────────────────────┬───────────────────┘
               │                       │
    ┌──────────▼──────────┐   ┌───────▼─────────────┐
    │  GREEN (active)     │   │  BLUE (new, standby) │
    │  lkfl-server:8080   │   │  lkfl-server:8081    │
    │  lkfl-proxy:8090    │   │  lkfl-proxy:8091     │
    │  lkfl-frontend:8086 │   │  (frontend — static) │
    └─────────────────────┘   └──────────────────────┘
```

### Процесс zero-downtime деплоя

```
Step 1: Pull new images (не влияет на running контейнеры)
  docker compose -f docker-compose.staging.yml pull lkfl-server lkfl-integration-proxy lkfl-frontend

Step 2: Run backward-compatible migration
  docker compose run --rm lkfl-migrate
  → migration добавляет колонки/таблицы, НЕ удаляет

Step 3: Start new service на alternate port
  docker compose -f docker-compose.blue.yml up -d
  → lkfl-server-new на порту 8081
  → lkfl-proxy-new на порту 8092

Step 4: Health check нового сервиса
  curl -sf http://127.0.0.1:8081/healthz
  → если ОК → продолжить
  → если НЕ ОК → kill blue → rollback (ADR-039)

Step 5: Nginx switch upstream (ATOMIC)
  # Переключить upstream с 8080 → 8081
  sed -i 's/server 127.0.0.1:8080/server 127.0.0.1:8081/' /etc/nginx/conf.d/lkfl.conf
  nginx -t && nginx -s reload
  → atomic reload: Nginx не теряет connections

Step 6: Graceful shutdown old service
  docker compose -f docker-compose.staging.yml stop lkfl-server lkfl-integration-proxy
  → graceful shutdown 30s (stop_grace_period: 35s)
  → существующие requests drain'ятся

Step 7: Cleanup
  docker compose -f docker-compose.staging.yml rm -f lkfl-server lkfl-integration-proxy
  → swap blue → green для следующего деплоя
```

### Migration strategy: Expand/Contract pattern

**Правило:** никогда не удалять колонку/таблицу в том же деплое, где новый код её использует.

```
Деплой V2 (Expand):
  ┌─────────────────────────────────────┐
  │ Migration: ADD COLUMN new_col       │ ← backward compatible
  │ Code: read old_col AND new_col      │ ← поддерживает оба
  │ Deploy: blue-green switch           │
  └─────────────────────────────────────┘

Деплой V3 (Contract):
  ┌─────────────────────────────────────┐
  │ Code: use only new_col              │ ← больше не читает old
  │ Deploy: blue-green switch           │
  │ Migration: DROP COLUMN old_col      │ ← безопасно, никто не читает
  └─────────────────────────────────────┘
```

| Операция | Безопасно для rollback? | Когда делать |
|----------|------------------------|-------------|
| `CREATE TABLE` | ✅ Да | Любой деплой |
| `ADD COLUMN col DEFAULT value` | ✅ Да | Любой деплой |
| `ADD COLUMN col NOT NULL DEFAULT value` | ✅ Да | Любой деплой |
| `ALTER COLUMN col TYPE` | ⚠️ Проверить | Только если old code не зависит от типа |
| `DROP COLUMN` | ❌ Нет | Только ПОСЛЕ deploy кода, который не использует |
| `DROP TABLE` | ❌ Нет | Только ПОСЛЕ deploy кода, который не использует |
| `RENAME COLUMN` | ❌ Нет | Заменить на expand/contract |

### Когда downtime НЕизбежен

| Ситуация | Downtime | Митигация |
|----------|----------|-----------|
| Breaking migration (rename column) | 3-5 мин | Planned maintenance window |
| Keycloak major upgrade | 30 сек OIDC disruption | Deprecate old realm + create new |
| PostgreSQL major version | 5-10 мин | Planned maintenance window |
| Redis persistence mode change | 1 мин (restart) | Off-peak hours |
| Nginx config breaking change | 0 (atomic reload) | Always atomic |

### Docker Compose реализация

**docker-compose.blue.yml** — override file для blue instances:

```yaml
# docker-compose.blue.yml
services:
  lkfl-server:
    container_name: lkfl-server-blue
    ports:
      - "127.0.0.1:8081:8080"  # alternate port
    # Остальное то же самое (image, env, depends_on)

  lkfl-integration-proxy:
    container_name: lkfl-proxy-blue
    ports:
      - "127.0.0.1:8092:8090"  # alternate port
      - "127.0.0.1:8093:8091"
```

**Nginx upstream config:**

```nginx
# /etc/nginx/conf.d/lkfl.conf
upstream lkfl_server {
    server 127.0.0.1:8080;  # green (active)
    # server 127.0.0.1:8081;  # blue (standby)
}

# После switch — swap:
upstream lkfl_server {
    # server 127.0.0.1:8080;  # green (draining)
    server 127.0.0.1:8081;    # blue (active)
}
```

### CI/CD интеграция

```yaml
# .github/workflows/build.yml — deploy-staging job (улучшенный)
- name: Blue-green deploy
  run: |
    # Pull new images
    docker compose -f docker-compose.staging.yml pull

    # Migration (backward-compatible only!)
    docker compose -f docker-compose.staging.yml run --rm lkfl-migrate

    # Start blue
    docker compose -f docker-compose.blue.yml up -d

    # Health check blue
    for i in $(seq 1 30); do
      if curl -sf http://127.0.0.1:8081/healthz; then
        echo "Blue healthy"
        break
      fi
      sleep 2
    done

    # Switch nginx upstream
    # (swap ports in nginx config + reload)

    # Stop green
    docker compose -f docker-compose.staging.yml stop lkfl-server lkfl-integration-proxy
```

## Consequences

### Положительные
- **Zero downtime** для backward-compatible деплоев
- **Nginx reload atomic** — нет потери connections
- **Graceful drain** — существующие requests завершаются (30s)
- **Migration safety** — expand/contract pattern предотвращает breaking changes
- **Rollback прост** — переключить nginx обратно на green

### Отрицательные
- **Extra infra:** второй набор контейнеров на alternate ports (~5 мин extra CPU/memory)
- **Migration discipline критична:** любая breaking migration ломает rollback
- **Сложнее deploy workflow:** 7 шагов вместо 3
- **Port management:** нужно отслеживать green vs blue ports

### Риски

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| Breaking migration в production | Средняя | Code review миграций + staging test |
| Blue не здоров → оба down | Низкая | Health check before switch |
| Port conflict | Низкая | Фиксированные alternate ports в compose |
| Nginx reload теряет connection | Очень низкая | Nginx reload atomic по design |

## Production notes

Для production с 100K+ сотрудников:
- **Kubernetes** → native rolling updates + readiness probes + pod disruption budgets
- **Load balancer** → HAProxy или cloud LB вместо Nginx upstream
- **Database replication** → read replicas для миграций на отдельном сервере
- **Canary deployment** → 10% traffic → 50% → 100% (не blue-green)

Docker Compose blue-green подходит для текущей инфраструктуры (один сервер, staging). Production — рассмотреть K8s.

## Related ADR

- **ADR-030:** CI/CD Pipeline — базовая архитектура деплоя
- **ADR-036:** Deploy Worker — webhook-based деплой
- **ADR-039:** Rollback Strategy — rollback при провале blue-green
