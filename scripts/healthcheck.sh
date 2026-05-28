#!/bin/bash
# LKFL — Staging Healthcheck Script
# Веха M22 — F1 Hardening
#
# Проверка всех сервисов staging стенда.
# Запуск: ./scripts/healthcheck.sh
#
# Выход: 0 = все OK, 1 = есть проблемы

set -euo pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS=0
FAIL=0

check() {
    local name="$1"
    local cmd="$2"

    printf "  %-35s " "$name"
    if eval "$cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}OK${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}FAIL${NC}"
        FAIL=$((FAIL + 1))
    fi
}

echo -e "${YELLOW}LKFL Staging Healthcheck${NC}"
echo "========================="
echo ""

# --- Backend ---
echo -e "${YELLOW}Backend${NC}"
check "lkfl-server healthz" "curl -sf http://localhost:8080/healthz"
echo ""

# --- Frontend ---
echo -e "${YELLOW}Frontend${NC}"
check "frontend (HTTP 80)" "curl -sf http://localhost:80/nginx-health"
check "frontend (HTTPS 443)" "curl -sfk https://localhost:443/"
echo ""

# --- Nginx ---
echo -e "${YELLOW}Nginx${NC}"
check "nginx health endpoint" "curl -sf http://localhost:80/nginx-health"
echo ""

# --- Database ---
echo -e "${YELLOW}Database${NC}"
check "PostgreSQL (pg_isready)" "docker exec staging-postgres pg_isready -U lkfl"
check "PostgreSQL (lkfl_platform DB)" "docker exec staging-postgres psql -U lkfl -d lkfl_platform -c 'SELECT 1' -t"
echo ""

# --- Redis ---
echo -e "${YELLOW}Cache${NC}"
check "Redis (ping)" "docker exec staging-redis redis-cli ping | grep -q PONG"
echo ""

# --- Keycloak ---
echo -e "${YELLOW}Identity${NC}"
check "Keycloak (admin console)" "curl -sf http://localhost:8081/admin/master/console/"
check "Keycloak (realms)" "curl -sf http://localhost:8081/realms/master"
echo ""

# --- Integration Proxy ---
echo -e "${YELLOW}Integration Proxy${NC}"
check "proxy healthz (HTTP 8091)" "curl -sf http://localhost:8091/healthz"
echo ""

# --- Monitoring ---
echo -e "${YELLOW}Monitoring${NC}"
check "Prometheus" "curl -sf http://localhost:9090/-/healthy"
check "Grafana" "curl -sf http://localhost:3000/api/health"
check "Loki" "curl -sf http://localhost:3100/ready"
echo ""

# --- Summary ---
echo "========================="
TOTAL=$((PASS + FAIL))
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC} (total: ${TOTAL})"
echo ""

if [ "$FAIL" -gt 0 ]; then
    echo -e "${RED}Some checks failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All checks passed!${NC}"
    exit 0
fi
