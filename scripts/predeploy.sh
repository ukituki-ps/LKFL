#!/bin/bash
# LKFL — Pre-deploy Validation Script
#
# Проверяет готовность к деплою:
#   - Локальные проверки (Go build, npm build, docker compose config)
#   - Deploy-worker API (GET /status на порт 9091)
#
# Использование:
#   ./scripts/predeploy.sh           # все проверки
#   ./scripts/predeploy.sh --quick   # только критичные проверки
#   ./scripts/predeploy.sh --local   # только локальные проверки

set -euo pipefail

# ============================================================================
# Конфигурация
# ============================================================================

DEPLOY_WORKER_URL="${DEPLOY_WORKER_URL:-http://serverAI:9091}"
BACKEND_DIR="backend"
FRONTEND_DIR="frontend"

# ============================================================================
# Цвета и вывод
# ============================================================================

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PASS=0
FAIL=0
WARN=0

check() {
    local name="$1"
    local cmd="$2"

    printf "  %-45s " "$name"
    if eval "$cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}✗ FAIL${NC}"
        FAIL=$((FAIL + 1))
    fi
}

warn() {
    local name="$1"
    local cmd="$2"

    printf "  %-45s " "$name"
    if eval "$cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "${YELLOW}⚠ WARN${NC}"
        WARN=$((WARN + 1))
    fi
}

# ============================================================================
# Локальные проверки
# ============================================================================

check_local() {
    echo -e "${CYAN}--- Локальные проверки ---${NC}"
    echo ""

    check "Go установлен" "which go > /dev/null 2>&1"
    check "Go модуль (go.mod)" "test -f ${BACKEND_DIR}/go.mod"
    check "Go сборка (lkfl-server)" "cd ${BACKEND_DIR} && go build -o /dev/null ./cmd/server/ 2>/dev/null"
    check "Go сборка (lkfl-proxy)" "cd ${BACKEND_DIR} && go build -o /dev/null ./cmd/integration-proxy/ 2>/dev/null"
    check "Go сборка (deploy-worker)" "cd ${BACKEND_DIR} && go build -o /dev/null ./cmd/deploy-worker/ 2>/dev/null"
    check "Go vet" "cd ${BACKEND_DIR} && go vet ./... 2>/dev/null"

    check "Node.js установлен" "which node > /dev/null 2>&1"
    check "npm установлен" "which npm > /dev/null 2>&1"
    check "package.json" "test -f ${FRONTEND_DIR}/package.json"
    warn "node_modules" "test -d ${FRONTEND_DIR}/node_modules"

    check "Docker установлен" "which docker > /dev/null 2>&1"
    check "Docker daemon" "docker info > /dev/null 2>&1"
    check "Docker Compose" "docker compose version > /dev/null 2>&1"

    check "Dockerfile.server" "test -f Dockerfile.server"
    check "Dockerfile.proxy" "test -f Dockerfile.proxy"
    check "Dockerfile.frontend" "test -f Dockerfile.frontend"
    check "Dockerfile.deploy-worker" "test -f Dockerfile.deploy-worker"
    check "docker-compose.dev.yml" "test -f docker-compose.dev.yml"
    check "docker-compose.staging.yml" "test -f docker-compose.staging.yml"
    check "Makefile" "test -f Makefile"
    check "Миграции" "test -d migrations"

    # Валидация compose файлов
    warn "docker-compose.dev.yml валидный" "docker compose -f docker-compose.dev.yml config > /dev/null 2>&1"
    warn "docker-compose.staging.yml валидный" "docker compose -f docker-compose.staging.yml config > /dev/null 2>&1"

    # Тесты (полный режим)
    if [ "${MODE:-full}" = "full" ]; then
        check "Go тесты (short)" "cd ${BACKEND_DIR} && go test ./... -short -count=1 2>/dev/null"
        warn "Frontend lint" "cd ${FRONTEND_DIR} && npx eslint src/ 2>/dev/null"
    fi

    echo ""
}

# ============================================================================
# Проверки deploy-worker API
# ============================================================================

check_deploy_worker() {
    echo -e "${CYAN}--- Deploy Worker (${DEPLOY_WORKER_URL}) ---${NC}"
    echo ""

    check "Deploy-worker доступен" "curl -sf ${DEPLOY_WORKER_URL}/healthz"
    check "GET /status работает" "curl -sf ${DEPLOY_WORKER_URL}/status"

    echo ""
}

# ============================================================================
# Main
# ============================================================================

MODE="${1:-full}"

echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  LKFL Pre-deploy Validation${NC}"
echo -e "${YELLOW}  Режим: ${MODE}${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

case "$MODE" in
    --quick)
        check_local 2>/dev/null | grep -E "(FAIL|WARN)" || true
        ;;
    --local)
        check_local
        ;;
    *)
        MODE="full"
        check_local
        check_deploy_worker
        ;;
esac

# ============================================================================
# Итог
# ============================================================================

echo -e "${YELLOW}========================================${NC}"
echo -e "  Результаты:"
echo -e "  ${GREEN}✓ ${PASS} OK${NC}"
echo -e "  ${YELLOW}⚠ ${WARN} Warning${NC}"
echo -e "  ${RED}✗ ${FAIL} Failed${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

if [ "$FAIL" -gt 0 ]; then
    echo -e "${RED}Нельзя деплоить — есть ошибки.${NC}"
    exit 1
elif [ "$WARN" -gt 0 ]; then
    echo -e "${YELLOW}Деплой возможен, но есть предупреждения.${NC}"
    exit 0
else
    echo -e "${GREEN}Всё чисто — готово к деплою!${NC}"
    exit 0
fi
