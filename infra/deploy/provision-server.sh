# Provision serverAi — one-time setup
# Run locally: bash infra/deploy/provision-server.sh serverAi
# Run on server: ssh serverAi 'bash -s' < infra/deploy/provision-server.sh
#
# Делает то, что НЕЛЬЗЯ через git:
#   - Docker volumes (persistent data)
#   - .env с секретами
#   - Разовая подготовка домашней директории
#
# Idempotent — безопасно перезапускать.
set -euo pipefail

LKFL_HOME="$HOME/lkfl"
mkdir -p "$LKFL_HOME"

echo "=== Provisioning serverAi for LKFL ==="
echo "  Home: $LKFL_HOME"

# ── Docker volumes ─────────────────────────────────────────
# Persistent data — survives container restarts
for vol in lkfl_postgres_data lkfl_redis_data; do
    if docker volume inspect "$vol" >/dev/null 2>&1; then
        echo "  ⏭️  Volume $vol exists"
    else
        docker volume create "$vol"
        echo "  ✅ Created volume $vol"
    fi
done

# ── .env check ─────────────────────────────────────────────
if [ -f "$LKFL_HOME/.env" ]; then
    echo "  ⏭️  $LKFL_HOME/.env exists"
else
    echo "  ⚠️  $LKFL_HOME/.env not found"
    echo "     Create it with: POSTGRES_PASSWORD, KEYCLOAK_ADMIN_PASSWORD,"
    echo "     REDIS_PASSWORD, JWT_SECRET"
    cat > "$LKFL_HOME/.env.example" <<'EOF'
# === Staging .env — serverAi ===
# Скопировать в .env, заполнить секреты

POSTGRES_PASSWORD=<generate-random>
REDIS_PASSWORD=<generate-random>
KEYCLOAK_ADMIN=kcadmin
KEYCLOAK_ADMIN_PASSWORD=<generate-random>
JWT_SECRET=<generate-random>

# Deploy config
IMAGE_TAG=latest
REGISTRY=ghcr.io
ORG=lkfl
EOF
    echo "  ✅ Created $LKFL_HOME/.env.example"
    echo "     → cp .env.example .env && edit secrets"
fi

echo ""
echo "=== Provisioning complete ==="
echo "  Home:     $LKFL_HOME"
echo "  Volumes:  lkfl_postgres_data, lkfl_redis_data"
echo "  Next:     GH Actions deploy.yml → SSH → docker compose up"
