#!/usr/bin/env bash
# dev-deploy.sh — ручной деплой на dev стенд (serverAi)
# ADR-043: Два стенда — dev (ручной) + staging (CI/CD)
#
# Запуск: infra/deploy/dev-deploy.sh [IMAGE_TAG] [BRANCH]
# Пример:  infra/deploy/dev-deploy.sh main-latest
#          infra/deploy/dev-deploy.sh main-a1b2c3d feature-x-login
# Без аргументов — pull origin/main, берёт последний IMAGE_TAG из .env
#
# Работает на serverAi (192.168.1.27) в директории LKFL-dev (git repo).
# Idempotent — безопасно перезапускать.
set -euo pipefail

# ── Параметры ───────────────────────────────────────────────
IMAGE_TAG="${1:-}"
BRANCH="${2:-main}"
COMPOSE_FILE="docker-compose.dev-server.yml"
ENV_FILE=".env.dev-server"
PROJECT_NAME="lkfl-dev"
DEV_DIR="/home/ukituki/LKFL-dev"

echo "============================================"
echo "  LKFL Dev Deploy — project.ukituki.tech"
echo "  BRANCH: $BRANCH"
echo "  IMAGE_TAG: ${IMAGE_TAG:-из .env}"
echo "============================================"

# ── Git pull ────────────────────────────────────────────────
cd "$DEV_DIR"

if [[ ! -d ".git" ]]; then
    echo "❌ $DEV_DIR не является git-репозиторием"
    exit 1
fi

echo ""
echo "🔄 Git pull origin/$BRANCH..."
CURRENT_HASH=$(git rev-parse --short HEAD)
git fetch origin "$BRANCH" 2>&1 | tail -3
git reset --hard "origin/$BRANCH" 2>&1 | tail -2
NEW_HASH=$(git rev-parse --short HEAD)

if [[ "$CURRENT_HASH" != "$NEW_HASH" ]]; then
    echo "  ✅ Обновлён: $CURRENT_HASH → $NEW_HASH"
else
    echo "  ⏭️  Уже актуален: $NEW_HASH"
fi

# ── Проверки ────────────────────────────────────────────────
if [[ ! -f "$DEV_DIR/$COMPOSE_FILE" ]]; then
    echo "❌ $COMPOSE_FILE не найден после git pull"
    exit 1
fi

if [[ ! -f "$DEV_DIR/$ENV_FILE" ]]; then
    echo "⚠️  $ENV_FILE не найден, создаю из example..."
    cp "$DEV_DIR/.env.dev-server.example" "$DEV_DIR/$ENV_FILE"
fi

# ── Обновить IMAGE_TAG в .env ──────────────────────────────
if [[ -n "$IMAGE_TAG" ]]; then
    if grep -q '^IMAGE_TAG=' "$ENV_FILE"; then
        sed -i "s|^IMAGE_TAG=.*|IMAGE_TAG=$IMAGE_TAG|" "$ENV_FILE"
    else
        echo "IMAGE_TAG=$IMAGE_TAG" >> "$ENV_FILE"
    fi
fi

export IMAGE_TAG=$(grep '^IMAGE_TAG=' "$ENV_FILE" | cut -d'=' -f2)
export GHCR_REGISTRY="${GHCR_REGISTRY:-ghcr.io/ukituki-ps/lkfl}"

COMPOSE="docker compose -f $COMPOSE_FILE --env-file $ENV_FILE -p $PROJECT_NAME"

echo ""
echo "📦 Проверка локальных образов..."
# Images уже кэшированы на serverAi (CI/CD push). Пытаемся pull, но не падаем.
$COMPOSE pull lkfl-server 2>&1 || echo "⚠️  Pull пропущен (образ $IMAGE_TAG уже локально)"

# ── Миграции ────────────────────────────────────────────────
echo ""
echo "🗄️  Миграции..."
MIGRATION_OK=false
for i in 1 2 3; do
    echo "  Попытка $i/3..."
    if timeout 90 $COMPOSE run --rm lkfl-migrate; then
        MIGRATION_OK=true
        break
    fi
    echo "  ❌ Попытка $i провалена, повтор через 10s..."
    sleep 10
done

if [[ "$MIGRATION_OK" != "true" ]]; then
    echo "❌ Миграции провалились после 3 попыток"
    $COMPOSE logs --tail=50 lkfl-migrate 2>/dev/null || true
    exit 1
fi

# ── Seed ────────────────────────────────────────────────────
echo ""
echo "🌱 Seed..."
timeout 60 $COMPOSE run --rm lkfl-seed 2>&1 || echo "⚠️  Seed пропущен (данные могут уже существовать)"

# ── Перезапуск сервисов ────────────────────────────────────
echo ""
echo "🔄 Перезапуск сервисов..."
$COMPOSE down --remove-orphans 2>&1 | tail -3
$COMPOSE up -d 2>&1 | tail -10

# ── Health check ────────────────────────────────────────────
echo ""
echo "🏥 Health check..."
HEALTH_URL="http://127.0.0.1:18082/healthz"
MAX_RETRIES=20
for attempt in $(seq 1 $MAX_RETRIES); do
    if curl -sf "$HEALTH_URL" >/dev/null 2>&1; then
        echo "  ✅ lkfl-server healthy (попытка $attempt)"
        break
    fi
    if [[ $attempt -eq $MAX_RETRIES ]]; then
        echo "  ❌ Health check провален после $MAX_RETRIES попыток"
        $COMPOSE logs --tail=50 lkfl-server
        exit 1
    fi
    echo "  ⏳ Ожидание... (попытка $attempt/$MAX_RETRIES)"
    sleep 5
done

# ── Проверка Keycloak ──────────────────────────────────────
echo ""
echo "🔑 Проверка Keycloak..."
KC_OK=false
for i in $(seq 1 10); do
    if $COMPOSE ps keycloak 2>/dev/null | grep -q healthy; then
        KC_OK=true
        echo "  ✅ Keycloak healthy"
        break
    fi
    echo "  ⏳ Keycloak загружается... (попытка $i/10)"
    sleep 5
done

# ── Cleanup ─────────────────────────────────────────────────
echo ""
echo "🧹 Очистка старых образов..."
docker image prune -f --filter "until=48h" 2>/dev/null || true

# ── Итог ───────────────────────────────────────────────────
echo ""
echo "============================================"
echo "  ✅ Dev deploy завершён!"
echo "  URL: https://project.ukituki.tech"
echo "  Backend: http://127.0.0.1:18082/healthz"
echo "  Keycloak: http://127.0.0.1:19083"
echo "  Frontend: http://127.0.0.1:18088"
echo "  IMAGE_TAG: $IMAGE_TAG"
echo "============================================"
