#!/usr/bin/env bash
# dev-deploy.sh — ручной деплой на dev стенд (serverAi)
# ADR-043: Два стенда — dev (ручной) + staging (CI/CD)
#
# Запуск: infra/deploy/dev-deploy.sh <IMAGE_TAG>
# Пример:  infra/deploy/dev-deploy.sh main-a1b2c3d
#          infra/deploy/dev-deploy.sh feature-x-login-abc1234
# Без аргумента — берёт последний main-latest
#
# Работает на serverAi (192.168.1.27). Не нужен SSH — скрипт выполняется
# локально на сервере.
#
# Idempotent — безопасно перезапускать.
set -euo pipefail

# ── Параметры ───────────────────────────────────────────────
IMAGE_TAG="${1:-main-latest}"
COMPOSE_FILE="docker-compose.dev-server.yml"
ENV_FILE=".env.dev-server"
PROJECT_NAME="lkfl-dev"
DEV_DIR="/home/ukituki/LKFL-dev"
COMPOSE="docker compose -f $COMPOSE_FILE --env-file $ENV_FILE -p $PROJECT_NAME"

echo "============================================"
echo "  LKFL Dev Deploy — project.ukituki.tech"
echo "  IMAGE_TAG: $IMAGE_TAG"
echo "  DIR: $DEV_DIR"
echo "============================================"

# ── Проверки ────────────────────────────────────────────────
if [[ ! -f "$DEV_DIR/$COMPOSE_FILE" ]]; then
    echo "❌ $COMPOSE_FILE не найден в $DEV_DIR"
    echo "   Синхронизируйте файлы: cp ~/LKFL/$COMPOSE_FILE $DEV_DIR/"
    exit 1
fi

if [[ ! -f "$DEV_DIR/$ENV_FILE" ]]; then
    echo "❌ $ENV_FILE не найден в $DEV_DIR"
    echo "   Создайте: cp $DEV_DIR/.env.dev-server.example $ENV_FILE"
    exit 1
fi

cd "$DEV_DIR"

# ── Обновить IMAGE_TAG в .env ──────────────────────────────
if grep -q '^IMAGE_TAG=' "$ENV_FILE"; then
    sed -i "s|^IMAGE_TAG=.*|IMAGE_TAG=$IMAGE_TAG|" "$ENV_FILE"
else
    echo "IMAGE_TAG=$IMAGE_TAG" >> "$ENV_FILE"
fi

export IMAGE_TAG
export GHCR_REGISTRY="${GHCR_REGISTRY:-ghcr.io/ukituki-ps/lkfl}"

echo ""
echo "📦 Pull образов из GHCR..."
$COMPOSE pull --ignore-buildable lkfl-server lkfl-integration-proxy lkfl-frontend 2>&1 || {
    echo "⚠️  Pull завершён с предупреждениями (возможно старые образы для proxy/frontend)"
}

# ── Миграции ────────────────────────────────────────────────
echo ""
echo "🗄️  Миграции..."
MIGRATION_OK=false
for i in 1 2 3; do
    echo "  Попытка $i/3..."
    if $COMPOSE run --rm lkfl-migrate; then
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
$COMPOSE run --rm lkfl-seed 2>&1 || echo "⚠️  Seed пропущен (данные могут уже существовать)"

# ── Перезапуск сервисов ────────────────────────────────────
echo ""
echo "🔄 Перезапуск сервисов..."
$COMPOSE down --remove-orphans
$COMPOSE up -d

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
echo "  Frontend: http://127.0.0.1:8088"
echo "  IMAGE_TAG: $IMAGE_TAG"
echo "============================================"
