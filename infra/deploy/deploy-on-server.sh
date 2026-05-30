#!/usr/bin/env bash
# deploy-on-server.sh — runs on serverAi via SSH from GitHub Actions
# Usage: ssh serverAi 'bash -s' < deploy-on-server.sh <IMAGE_TAG>
set -euo pipefail

IMAGE_TAG="${1:?Usage: deploy-on-server.sh <image-tag>}"
LKFL_HOME="$HOME/lkfl"
COMPOSE_FILE="$LKFL_HOME/docker-compose.prod.yml"
ENV_FILE="$LKFL_HOME/.env"

echo "🚀 Deploy LKFL → serverAi (tag: $IMAGE_TAG)"

# ── Pull latest repo ───────────────────────────────────────
if [ -d "$LKFL_HOME/.git" ]; then
    cd "$LKFL_HOME"
    git fetch origin main
    git reset --hard origin/main
else
    mkdir -p "$LKFL_HOME"
    git clone "${LKFL_GIT_REPO:-git@github.com:lkfl/lkfl.git}" "$LKFL_HOME"
    cd "$LKFL_HOME"
fi

# ── Ensure .env exists ─────────────────────────────────────
if [ ! -f "$ENV_FILE" ]; then
    echo "⚠️  .env not found at $ENV_FILE"
    echo "   Run provision-server.sh first to set up .env"
    exit 1
fi

# ── Export IMAGE_TAG ───────────────────────────────────────
export IMAGE_TAG
export REGISTRY="${REGISTRY:-ghcr.io}"
export ORG="${ORG:-lkfl}"

# ── Pull images ────────────────────────────────────────────
echo "📦 Pulling images..."
docker compose -f "$COMPOSE_FILE" pull --policy missing

# ── Start services ─────────────────────────────────────────
echo "🟢 Starting services..."
docker compose -f "$COMPOSE_FILE" up -d

# ── Health check ───────────────────────────────────────────
echo "🏥 Health check..."
MAX_RETRIES=30
for i in $(seq 1 $MAX_RETRIES); do
    if curl -sf http://localhost:80/healthz >/dev/null 2>&1; then
        echo "✅ lkfl-server healthy (attempt $i)"
        break
    fi
    if [ "$i" -eq "$MAX_RETRIES" ]; then
        echo "❌ Health check failed after $MAX_RETRIES attempts"
        docker compose -f "$COMPOSE_FILE" logs --tail=50
        exit 1
    fi
    sleep 2
done

# ── Cleanup ────────────────────────────────────────────────
echo "🧹 Pruning old images..."
docker image prune -f --filter "until=24h"

echo "✅ Deploy complete: $IMAGE_TAG"
echo "   URL: http://$(hostname -I | awk '{print $1}')"
