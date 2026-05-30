#!/bin/bash
# LKFL — Deploy Script (выполняется внутри deploy-worker на serverAI)
#
# Новый пайплайн: pull из GHCR → миграции → up → seed → healthcheck
# Вызывается deploy-worker (cmd/deploy-worker) при получении webhook.
#
# Использование:
#   ./scripts/deploy.sh              # полный деплой
#   ./scripts/deploy.sh --dry-run    # только pull образов
#   ./scripts/deploy.sh --health     # только healthcheck
#   ./scripts/deploy.sh --rollback   # откат к предыдущему IMAGE_TAG
#
# Предварительные условия:
#   - Docker socket доступен (/var/run/docker.sock)
#   - .env.staging содержит IMAGE_TAG, GHCR_TOKEN, DB_DSN
#   - docker-compose.staging.yml присутствует

set -euo pipefail

# ============================================================================
# Конфигурация
# ============================================================================

COMPOSE_FILE="docker-compose.staging.yml"
ENV_FILE=".env.staging"
COMPOSE="docker compose -f ${COMPOSE_FILE} --env-file ${ENV_FILE}"

# IMAGE_TAG из .env.staging
IMAGE_TAG="${IMAGE_TAG:-main-latest}"
GHCR_REGISTRY="${GHCR_REGISTRY:-ghcr.io/ukituki-ps/lkfl}"

# ============================================================================
# Цвета и вывод
# ============================================================================

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

LOG_FILE="/tmp/lkfl-deploy-$(date +%Y%m%d-%H%M%S).log"

log() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $*"
    echo -e "$msg"
    echo "$msg" >> "$LOG_FILE"
}

log_ok() { log -e "${GREEN}✓${NC} $*"; }
log_err() { log -e "${RED}✗${NC} $*"; }
log_warn() { log -e "${YELLOW}⚠${NC} $*"; }
log_step() { log -e "${CYAN}▸${NC} $*"; }

# ============================================================================
# Утилиты
# ============================================================================

# Ожидание HTTP endpoint
wait_for_http() {
    local url="$1"
    local timeout="${2:-60}"
    local interval=2
    local elapsed=0

    while [ $elapsed -lt $timeout ]; do
        if curl -sf "$url" > /dev/null 2>&1; then
            return 0
        fi
        sleep $interval
        elapsed=$((elapsed + interval))
    done

    return 1
}

# ============================================================================
# Шаги деплоя
# ============================================================================

# 0. GHCR login
step_ghcr_login() {
    log_step "Login в GHCR..."

    if [ -n "${GHCR_TOKEN:-}" ]; then
        echo "$GHCR_TOKEN" | docker login ghcr.io -u ukituki --password-stdin 2>/dev/null
        log_ok "GHCR login OK"
    else
        log_warn "GHCR_TOKEN не установлен, пропускаю login"
    fi
}

# 1. Pull образов из GHCR
step_pull() {
    log_step "Pull образов из GHCR (tag: ${IMAGE_TAG})..."

    if $COMPOSE pull --ignore-buildable 2>&1 | tee -a "$LOG_FILE"; then
        log_ok "Образы pulled"
    else
        log_err "Ошибка pull образов"
        return 1
    fi
}

# 2. Миграции БД (idempotent)
step_migrate() {
    log_step "Применение миграций..."

    if $COMPOSE run --rm lkfl-migrate 2>&1 | tee -a "$LOG_FILE"; then
        log_ok "Миграции применены"
    else
        log_warn "Миграции не применены (возможно, уже актуальны)"
    fi
}

# 3. Запуск сервисов
step_up() {
    log_step "Запуск сервисов..."

    if $COMPOSE up -d 2>&1 | tee -a "$LOG_FILE"; then
        log_ok "Сервисы запущены"
    else
        log_err "Ошибка запуска сервисов"
        return 1
    fi
}

# 4. Seed данных (idempotent)
step_seed() {
    log_step "Загрузка seed данных..."

    if $COMPOSE run --rm lkfl-seed 2>&1 | tee -a "$LOG_FILE"; then
        log_ok "Seed данные загружены"
    else
        log_warn "Seed не загружен (данные могут уже существовать)"
    fi
}

# 5. Healthcheck
step_healthcheck() {
    log_step "Healthcheck..."

    # Ожидание готовности lkfl-server
    if wait_for_http "http://lkfl-server:8080/healthz" 60; then
        log_ok "lkfl-server healthcheck OK"
    else
        log_warn "lkfl-server healthcheck не прошёл (может стабилизироваться позже)"
    fi

    # Проверка nginx
    if wait_for_http "http://localhost:80/nginx-health" 30; then
        log_ok "nginx healthcheck OK"
    else
        log_warn "nginx healthcheck не прошёл"
    fi
}

# Статус
show_status() {
    log_step "Статус сервисов:"
    echo ""
    $COMPOSE ps 2>/dev/null || true
    echo ""
    log_step "Последние логи lkfl-server:"
    $COMPOSE logs --tail=20 lkfl-server 2>&1 | tail -20
}

# Роллбэк
do_rollback() {
    local previous_tag=""

    # 1. Попытка прочитать из .deploy-previous-tag (сохраняется deploy-worker)
    if [ -f ".deploy-previous-tag" ]; then
        previous_tag=$(cat .deploy-previous-tag | tr -d '[:space:]')
    fi

    # 2. Fallback: текущий IMAGE_TAG из .env.staging
    if [ -z "$previous_tag" ]; then
        previous_tag=$(grep -m1 '^IMAGE_TAG=' "$ENV_FILE" 2>/dev/null | cut -d= -f2- || echo "")
    fi

    if [ -z "$previous_tag" ]; then
        log_err "Нет предыдущего IMAGE_TAG для роллбэка"
        return 1
    fi

    log_warn "Роллбэк к IMAGE_TAG=${previous_tag}"
    export IMAGE_TAG="$previous_tag"

    step_ghcr_login
    step_pull || return 1
    step_migrate
    step_up || return 1
    step_seed
    step_healthcheck

    log_ok "Роллбэк завершён (IMAGE_TAG=${previous_tag})"
}

# ============================================================================
# Main
# ============================================================================

main() {
    local mode="${1:-full}"

    echo ""
    log -e "${YELLOW}========================================${NC}"
    log -e "${YELLOW}  LKFL Deploy — serverAI${NC}"
    log -e "${YELLOW}  IMAGE_TAG: ${IMAGE_TAG}${NC}"
    log -e "${YELLOW}  Режим: ${mode}${NC}"
    log -e "${YELLOW}========================================${NC}"
    echo ""

    case "$mode" in
        --dry-run)
            step_ghcr_login
            step_pull
            log_ok "Dry-run завершён. Образы pulled, сервер не перезапущен."
            ;;
        --health)
            step_healthcheck
            show_status
            ;;
        --rollback)
            do_rollback
            ;;
        --help|-h)
            head -15 "$0" | tail -10
            exit 0
            ;;
        *)
            step_ghcr_login
            step_pull || exit 1
            step_migrate
            step_up || exit 1
            step_seed
            step_healthcheck
            show_status

            echo ""
            log -e "${GREEN}========================================${NC}"
            log -e "${GREEN}  Деплой завершён!${NC}"
            log -e "${GREEN}  IMAGE_TAG: ${IMAGE_TAG}${NC}"
            log -e "${GREEN}  URL: https://dev.april.ukituki.tech${NC}"
            log -e "${GREEN}========================================${NC}"
            ;;
    esac
}

main "$@"
