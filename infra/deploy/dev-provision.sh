#!/usr/bin/env bash
# dev-provision.sh — первичная настройка dev стенда на serverAi
#
# Запуск один раз при первом развёртывании. Idempotent.
#
# Запуск локально (копирует файлы на serverAi):
#   ssh serverAi 'bash -s' < infra/deploy/dev-provision.sh
#
# Запуск напрямую на serverAi:
#   bash infra/deploy/dev-provision.sh
set -euo pipefail

DEV_DIR="/home/ukituki/LKFL-dev"
REPO_DIR="${LKFL_REPO:-/home/ukituki/LKFL}"

echo "=== Provisioning dev стенд на serverAi ==="
echo "  Dev dir: $DEV_DIR"
echo "  Repo: $REPO_DIR"

# ── Создать директорию ─────────────────────────────────────
mkdir -p "$DEV_DIR"

# ── Копировать compose файл ────────────────────────────────
if [[ -f "$REPO_DIR/docker-compose.dev-server.yml" ]]; then
    cp "$REPO_DIR/docker-compose.dev-server.yml" "$DEV_DIR/"
    echo "  ✅ docker-compose.dev-server.yml → $DEV_DIR/"
else
    echo "  ⚠️  docker-compose.dev-server.yml не найден в $REPO_DIR"
    echo "     Скопируйте вручную: cp docker-compose.dev-server.yml $DEV_DIR/"
fi

# ── Копировать .env.example ────────────────────────────────
if [[ -f "$REPO_DIR/.env.dev-server.example" ]]; then
    if [[ ! -f "$DEV_DIR/.env.dev-server" ]]; then
        cp "$REPO_DIR/.env.dev-server.example" "$DEV_DIR/.env.dev-server"
        echo "  ✅ .env.dev-server создан из example"
        echo "     → Отредактируйте секреты: $DEV_DIR/.env.dev-server"
    else
        echo "  ⏭️  .env.dev-server уже существует"
    fi
else
    echo "  ⚠️  .env.dev-server.example не найден"
fi

# ── Копировать infra (keycloak realm, postgres init) ───────
mkdir -p "$DEV_DIR/infra"

for sub_dir in keycloak postgres; do
    if [[ -d "$REPO_DIR/infra/$sub_dir" ]]; then
        rm -rf "$DEV_DIR/infra/$sub_dir"
        cp -r "$REPO_DIR/infra/$sub_dir" "$DEV_DIR/infra/$sub_dir"
        echo "  ✅ infra/$sub_dir/ → $DEV_DIR/infra/$sub_dir/"
    fi
done

# ── Nginx serverAi — internal proxy (port 18002) ──────────
if [[ -f "$REPO_DIR/infra/nginx/serverAi-dev.conf" ]]; then
    echo ""
    echo "  📋 Nginx serverAi internal (port 18002):"
    echo "     sudo cp $REPO_DIR/infra/nginx/serverAi-dev.conf /etc/nginx/sites-available/lkfl-dev-internal.conf"
    echo "     sudo ln -sf /etc/nginx/sites-available/lkfl-dev-internal.conf /etc/nginx/sites-enabled/"
    echo "     sudo nginx -t && sudo nginx -s reload"
    echo ""
    # Попробовать сделать автоматически
    if sudo -n true 2>/dev/null; then
        sudo cp "$REPO_DIR/infra/nginx/serverAi-dev.conf" /etc/nginx/sites-available/lkfl-dev-internal.conf
        sudo ln -sf /etc/nginx/sites-available/lkfl-dev-internal.conf /etc/nginx/sites-enabled/
        sudo nginx -t && sudo nginx -s reload
        echo "  ✅ Nginx serverAi обновлён (port 18002)"
    fi
fi

# ── Docker volumes ─────────────────────────────────────────
echo ""
echo "  Docker volumes:"
for vol in dev_pg_data dev_redis_data dev_prometheus_data dev_grafana_data dev_loki_data; do
    if docker volume inspect "$vol" >/dev/null 2>&1; then
        echo "    ⏭️  $vol exists"
    else
        echo "    ✅ Created $vol"
    fi
done

# ── Docker networks ────────────────────────────────────────
echo ""
echo "  Docker networks:"
for net in lkfl_backend_dev lkfl_frontend_dev; do
    if docker network inspect "$net" >/dev/null 2>&1; then
        echo "    ⏭️  $net exists"
    else
        docker network create "$net" >/dev/null 2>&1
        echo "    ✅ Created $net"
    fi
done

echo ""
echo "=== Provisioning завершён ==="
echo ""
echo "  Следующие шаги:"
echo "  1. Отредактируйте $DEV_DIR/.env.dev-server (секреты)"
echo "  2. Установите nginx serverPr01: infra/nginx/serverPr01-dev.conf"
echo "  3. Первый деплой: bash infra/deploy/dev-deploy.sh main-latest"
echo ""
echo "  URL: https://project.ukituki.tech"
