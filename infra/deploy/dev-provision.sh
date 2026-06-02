#!/usr/bin/env bash
# dev-provision.sh — первичная настройка dev стенда на serverAi (git workflow)
#
# Запуск один раз при первом развёртывании. Idempotent.
#
# Запуск на serverAi:
#   ssh serverAi 'bash infra/deploy/dev-provision.sh'
set -euo pipefail

DEV_DIR="/home/ukituki/LKFL-dev"
REPO_URL="git@github.com:ukituki-ps/LKFL.git"
BRANCH="${1:-main}"

echo "=== Provisioning dev стенд (git workflow) ==="
echo "  Dev dir: $DEV_DIR"
echo "  Repo: $REPO_URL"
echo "  Branch: $BRANCH"

# ── Git clone или update ───────────────────────────────────
if [[ -d "$DEV_DIR/.git" ]]; then
    echo ""
    echo "  ⏭️  $DEV_DIR уже является git-репозиторием"
    cd "$DEV_DIR"
    git fetch origin "$BRANCH" 2>&1 | tail -3
    git reset --hard "origin/$BRANCH" 2>&1 | tail -2
else
    if [[ -d "$DEV_DIR" ]]; then
        echo ""
        echo "  ⚠️  $DEV_DIR существует, но не git — удаляю и клонирую заново"
        rm -rf "$DEV_DIR"
    fi

    mkdir -p "$(dirname "$DEV_DIR")"
    cd "$(dirname "$DEV_DIR")"
    git clone --branch "$BRANCH" "$REPO_URL" "LKFL-dev" 2>&1 | tail -3
fi

# ── .env.dev-server из example (если нет) ──────────────────
if [[ ! -f "$DEV_DIR/.env.dev-server" ]]; then
    cp "$DEV_DIR/.env.dev-server.example" "$DEV_DIR/.env.dev-server"
    echo "  ✅ .env.dev-server создан из example"
else
    echo "  ⏭️  .env.dev-server уже существует"
fi

# ── Docker volumes ─────────────────────────────────────────
echo ""
echo "  Docker volumes:"
for vol in dev_pg_data dev_redis_data; do
    if docker volume inspect "$vol" >/dev/null 2>&1; then
        echo "    ⏭️  $vol exists"
    else
        docker volume create "$vol" 2>/dev/null
        echo "    ✅ Created $vol"
    fi
done

echo ""
echo "=== Provisioning завершён ==="
echo ""
echo "  Следующие шаги:"
echo "  1. Отредактируйте секреты: nano $DEV_DIR/.env.dev-server"
echo "  2. Первый деплой: cd $DEV_DIR && bash infra/deploy/dev-deploy.sh main-latest"
echo ""
echo "  URL: https://project.ukituki.tech"
