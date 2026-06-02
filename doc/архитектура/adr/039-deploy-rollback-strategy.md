# ADR-039: Стратегия отката при деплое (Deploy Rollback Strategy)

| Поле     | Значение |
|----------|----------|
| Status   | Accepted |
| Date     | 2026-06-02 |
| Веха     | F1 Hardening |
| Авторы   | architect-lkfl |
| Related  | ADR-030, ADR-036, ADR-037, ADR-038 |

## Context

Деплой на staging (serverAI) осуществляется через Deploy Worker (ADR-036): GitHub Actions → webhook POST `/deploy` → Deploy Worker выполняет `docker compose down → pull → migrate → seed → up → health check`.

**Проблема:** при провале деплоя сервис недоступен, автоматического rollback нет. Deploy Worker имеет `/rollback` endpoint (ADR-036), но стратегия отката не определена:

- Какой образ считать «предыдущим»?
- Что делать если миграции БД отработали частично?
- Сколько образов хранить в GHCR для rollback?
- Как сохранить состояние перед деплоем?

**Сценарии отказа:**

| Сценарий | Что сломалось | Время восстановления без rollback |
|----------|---------------|-----------------------------------|
| Health check провалился | lkfl-server не отвечает | Ручной rollback + compose restart |
| Migration провалилась | БД в неконсистентном состоянии | Ручной fix БД + rollback образа |
| Seed провалился | Отсутствуют reference-данные | Повтор seed + rollback |
| Panic на startup | Бинарник упал при старте | Быстрый rollback образа |
| Фронтенд белый экран | Breaking change в API/frontend | Быстрый rollback образа |

## Decision

### Стратегия отката

```
Деплой (успех):
  ┌─────────────────────────────────────────────────┐
  │ 1. Backup .env.staging → .env.staging.backup    │
  │ 2. docker compose down                          │
  │ 3. docker compose pull (новые образы)            │
  │ 4. migrate (one-shot)                           │
  │ 5. seed (one-shot)                              │
  │ 6. docker compose up -d                         │
  │ 7. Health check /healthz (30s timeout)          │
  │ 8. Сохранить IMAGE_TAG как previous_tag          │
  └─────────────────────────────────────────────────┘

Деплой (провал → rollback):
  ┌─────────────────────────────────────────────────┐
  │ 1. docker compose down                          │
  │ 2. Подставить previous_tag в compose            │
  │ 3. docker compose pull (старые образы)           │
  │ 4. docker compose up -d (без migrate/seed!)     │
  │ 5. Health check /healthz                        │
  │ 6. Alert: manual DB review required?             │
  └─────────────────────────────────────────────────┘
```

### 1. Manual rollback

Замена `IMAGE_TAG` в `.env.staging` на предыдущий тег + compose down/up:

```bash
# Сохранить текущий тег
cp .env.staging .env.staging.backup

# Откат
sed -i 's/IMAGE_TAG=main-a1b2c3d/IMAGE_TAG=main-xyz7890/' .env.staging
docker compose -f docker-compose.staging.yml down
docker compose -f docker-compose.staging.yml pull
docker compose -f docker-compose.staging.yml up -d
```

**Время:** ~2 мин (без migration).

### 2. Auto rollback в CI

При провале health check в `deploy.yml` — автоматически переключить на предыдущий tag:

```yaml
# .github/workflows/deploy.yml — health check с auto-rollback
- name: Health check
  id: health
  run: |
    for i in $(seq 1 30); do
      if curl -sf https://dev.april.ukituki.tech/healthz; then
        echo "healthy=true" >> $GITHUB_OUTPUT
        exit 0
      fi
      sleep 2
    done
    echo "healthy=false" >> $GITHUB_OUTPUT

- name: Auto rollback on health check failure
  if: steps.health.outputs.healthy == 'false'
  run: |
    curl -X POST https://dev.april.ukituki.tech:9092/rollback \
      -H "Authorization: Bearer ${{ secrets.DEPLOY_TOKEN }}"
```

### 3. Image retention

Хранить последние 10 тегов на каждый сервис в GHCR:

| Сервис | Image | Retention |
|--------|-------|-----------|
| server | `ghcr.io/ukituki-ps/lkfl/server` | последние 10 тегов |
| proxy | `ghcr.io/ukituki-ps/lkfl/proxy` | последние 10 тегов |
| frontend | `ghcr.io/ukituki-ps/lkfl/frontend` | последние 10 тегов |
| deploy-worker | `ghcr.io/ukituki-ps/lkfl/deploy-worker` | latest + previous |

Очистка старых тегов — cron job в GitHub Actions (еженедельно):

```yaml
# .github/workflows/cleanup.yml
- name: Clean old GHCR tags
  run: |
    # Удалить теги старше 10 последних для каждого сервиса
    gh api --method DELETE /orgs/ukituki-ps/packages/container/lkfl-server/version/{digest}
```

**Объём:** 10 тегов × 3 сервиса × ~70MB = ~2.1 GB в GHCR.

### 4. Migration rollback

**НЕТ backward migrations.** Миграции только forward.

| Сценарий | Действие |
|----------|----------|
| Migration провалилась на середине | Ручной fix БД + rollback образа без повторной migration |
| Migration создала новую таблицу, но следующий шаг упал | Оставить таблицу, rollback образа — старая версия сервера не знает о новой таблице |
| Migration изменила существующую таблицу (ALTER) | **Критично:** требуется совместимость backward — старая версия сервера работает с новой схемой |

**Правило backward compatibility:** каждая migration должна быть backward-compatible со старым кодом сервера. Это означает:

- ✅ Добавление новой таблицы — безопасно
- ✅ Добавление новой колонки с DEFAULT — безопасно
- ✅ Добавление NOT NULL колонки с DEFAULT — безопасно
- ⚠️ Изменение типа колонки — требует проверки
- ❌ Удаление колонки — НЕ безопасно (старый код её использует)
- ❌ Удаление таблицы — НЕ безопасно

**Процедура при провале migration:**

```bash
# 1. Остановить всё
docker compose -f docker-compose.staging.yml down

# 2. Подключиться к БД для диагностики
docker compose -f docker-compose.staging.yml run --rm lkfl-postgres psql \
  -U lkfl -d lkfl_platform -c "\dt"

# 3. Ручной fix если нужно (пример — удаление частично применённой migration)
# psql -c "DELETE FROM schema_migrations WHERE version = '2026060200';"

# 4. Rollback образа БЕЗ migration
sed -i 's/IMAGE_TAG=main-bad/IMAGE_TAG=main-good/' .env.staging
docker compose -f docker-compose.staging.yml pull
docker compose -f docker-compose.staging.yml up -d
# Не запускать migrate/seed — они уже были в предыдущем деплое
```

### 5. Checkpoints

Перед каждым деплоем сохранять состояние:

```bash
# В Deploy Worker, перед deploy
cp .env.staging .env.staging.backup.$(date +%Y%m%d%H%M%S)
# Сохранить список запущенных контейнеров и их образы
docker ps --format "{{.Names}}: {{.Image}}: {{.Status}}" > /tmp/deploy-checkpoint.txt
# Сохранить текущий IMAGE_TAG
grep IMAGE_TAG .env.staging > /tmp/image-tag-backup.txt
```

Хранить последние 5 checkpoint'ов.

### Deploy Worker — расширенный `/rollback` endpoint

```go
// POST /rollback
type RollbackRequest struct {
    Service  string `json:"service"`  // "all", "server", "proxy", "frontend"
    Target   string `json:"target"`   // "previous" или конкретный IMAGE_TAG
    SkipMigrate bool `json:"skip_migrate"` // true по умолчанию
}
```

**Логика:**

1. Найти previous_tag из checkpoint
2. Подставить в compose
3. `docker compose down` → `pull` → `up -d` (без migrate/seed)
4. Health check
5. Alert если health check провалился (требует ручного вмешательства)

## Consequences

### Положительные

- **Rollback time:** ~2 мин (без migration) — приемлемо для staging
- **Auto rollback:** деплой не оставляет сервис в «сгоревшем» состоянии
- **Image retention:** 10 тегов — достаточно для отката на любую недавнюю версию
- **Checkpoints:** всегда есть точка восстановления
- **Migration safety:** backward compatibility правило предотвращает 90% проблем

### Отрицательные

- **Migration risk:** partial migration → требует manual DB fix (нет автоматического rollback БД)
- **Image storage:** ~2-3 GB в GHCR для 10 тегов × 3 сервиса
- **No zero-downtime:** между down и up есть окно недоступности (~30-60s)
- **Manual DB review:** при провале migration — всегда нужен инженер

### Риски

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| Forward migration + backward incompatible schema | Средняя | Code review миграций, backward compatibility правило |
| Роллбэк не работает (старый образ не запускается на новой схеме) | Низкая | Тестировать rollback на staging перед production |
| Потеря previous_tag | Низкая | Checkpoint + GHCR retention + git history .env.staging.example |

## Production considerations

Для production стратегия усложняется:

- **Zero-downtime deploy:** blue-green или rolling update (не сейчас)
- **WAL archiving:** для point-in-time recovery (RPO < 5 мин)
- **Automated rollback:** auto на production только при health check failure, с уведомлением в Slack/Telegram
- **Backup перед deploy:** pg_dump перед каждой миграцией в production

## Related ADR

- **ADR-030:** CI/CD Pipeline — базовая архитектура CI/CD, registry ghcr.io, deploy workflow
- **ADR-036:** Deploy Worker — webhook-based деплой, `/rollback` endpoint, architecture
- **ADR-037:** Keycloak reverse proxy — влияет на rollback (nginx config не меняется при rollback образа)
- **ADR-038:** Переезд staging — serverAI как единый хост, контекст инфраструктуры
