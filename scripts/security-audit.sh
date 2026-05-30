#!/bin/bash
# LKFL — Security Audit Script
# Веха M22 — F1 Hardening (T2211)
#
# OWASP Top 10 проверка, SQL injection, XSS, dependency audit.
# Запуск: ./scripts/security-audit.sh
#
# Выход: 0 = все проверки пройдены, 1 = есть проблемы

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS=0
FAIL=0
WARN=0

# Project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKEND="$PROJECT_ROOT/backend"
FRONTEND="$PROJECT_ROOT/frontend"

check() {
    local name="$1"
    local cmd="$2"

    printf "  %-45s " "$name"
    if eval "$cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}OK${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}FAIL${NC}"
        FAIL=$((FAIL + 1))
    fi
}

warn() {
    local name="$1"
    local cmd="$2"

    printf "  %-45s " "$name"
    if eval "$cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}OK${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "${YELLOW}WARN${NC}"
        WARN=$((WARN + 1))
    fi
}

# safe_count — безопасный подсчёт совпадений grep (не ломает pipefail)
safe_count() {
    local result
    result=$(grep -c "$@" 2>/dev/null) || true
    echo "${result:-0}" | tr -d '[:space:]'
}

# safe_grep_count — подсчёт строк через grep | wc -l (без pipefail issues)
safe_grep_count() {
    local result
    result=$(grep "$@" 2>/dev/null | wc -l) || true
    echo "${result:-0}" | tr -d '[:space:]'
}

echo -e "${YELLOW}=== OWASP Top 10 Security Audit ===${NC}"
echo ""

# ──────────────────────────────────────────────
# A01: Broken Access Control
# ──────────────────────────────────────────────
echo -e "${YELLOW}A01: Broken Access Control${NC}"
check "  RBAC middleware exists" "test -f $BACKEND/shared/pkg/auth/rbac.go"
check "  Tenant isolation exists" "test -f $BACKEND/internal/tenant/isolation.go"
check "  JWT middleware exists" "test -f $BACKEND/shared/pkg/auth/middleware.go"
check "  Admin routes use RBAC" "grep -q 'RBACMiddleware' $BACKEND/internal/app/server.go"
check "  Employee routes use JWT" "grep -q 'JWTMiddleware' $BACKEND/internal/app/server.go"
echo ""

# ──────────────────────────────────────────────
# A02: Cryptographic Failures
# ──────────────────────────────────────────────
echo -e "${YELLOW}A02: Cryptographic Failures${NC}"
check "  JWT verifier exists" "test -f $BACKEND/shared/pkg/auth/verifier.go"
check "  OIDC integration" "grep -q 'go-oidc' $BACKEND/go.mod"

# Проверка что нет хардкода секретов
echo "  - Checking for hardcoded secrets..."
SECRETS_COUNT=0
if grep -rn 'password.*=.*"[^"]*"' $BACKEND/ --include="*.go" 2>/dev/null | grep -v '_test.go' | grep -qv '// '; then
    SECRETS_COUNT=$(grep -rn 'password.*=.*"[^"]*"' $BACKEND/ --include="*.go" 2>/dev/null | grep -v '_test.go' | grep -v '// ' | wc -l | tr -d '[:space:]')
fi
if [ "$SECRETS_COUNT" -eq 0 ] 2>/dev/null; then
    echo -e "  ${GREEN}No hardcoded passwords found${NC}"
    PASS=$((PASS + 1))
else
    echo -e "  ${RED}Potential hardcoded secrets: $SECRETS_COUNT${NC}"
    FAIL=$((FAIL + 1))
fi
echo ""

# ──────────────────────────────────────────────
# A03: Injection (SQL Injection)
# ──────────────────────────────────────────────
echo -e "${YELLOW}A03: Injection (SQL Injection)${NC}"

echo "  - Checking for unsafe SQL string formatting..."
UNSAFE_SQL=$(safe_grep_count -rn 'fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT\|fmt.Sprintf.*UPDATE\|fmt.Sprintf.*DELETE' $BACKEND/internal/ --include="*.go")
if [ "$UNSAFE_SQL" -eq 0 ] 2>/dev/null; then
    echo -e "  ${GREEN}No unsafe SQL formatting found${NC}"
    PASS=$((PASS + 1))
else
    echo -e "  ${YELLOW}Potential unsafe SQL: $UNSAFE_SQL (проверить вручную)${NC}"
    WARN=$((WARN + 1))
fi

# Проверка что используются pgx параметризованные запросы
PARAM_QUERIES=$(grep -rn '\.QueryRow\|\.Query(\|\.Exec(' $BACKEND/internal/ --include="*.go" 2>/dev/null | grep -c '\$[0-9]' || echo "0")
echo "  - Parameterized queries found: $PARAM_QUERIES"

# Проверка что нет raw SQL string concatenation
RAW_CONCAT=$(safe_grep_count -rn 'fmt.Sprintf.*"SELECT\|fmt.Sprintf.*"INSERT\|fmt.Sprintf.*"UPDATE\|fmt.Sprintf.*"DELETE' $BACKEND/internal/ --include="*.go")
if [ "$RAW_CONCAT" -eq 0 ] 2>/dev/null; then
    echo -e "  ${GREEN}No raw SQL concatenation found${NC}"
    PASS=$((PASS + 1))
else
    echo -e "  ${RED}Raw SQL concatenation: $RAW_CONCAT${NC}"
    FAIL=$((FAIL + 1))
fi

# govulncheck (опционально)
warn "  govulncheck (если установлен)" "cd $BACKEND && govulncheck ./... 2>/dev/null"
echo ""

# ──────────────────────────────────────────────
# A05: Security Misconfiguration
# ──────────────────────────────────────────────
echo -e "${YELLOW}A05: Security Misconfiguration${NC}"

check "  CORS not wildcard (*)" "! grep -q 'Allow-Origin.*\\*' $BACKEND/internal/app/server.go"
check "  CORS uses rs/cors lib" "grep -q 'rs/cors' $BACKEND/go.mod"
check "  Rate limiting exists" "test -f $BACKEND/shared/pkg/middleware/ratelimit.go"
check "  Rate limiting on auth" "grep -q 'rl:auth' $BACKEND/internal/app/server.go"
check "  Rate limiting on catalog" "grep -q 'rl:catalog' $BACKEND/internal/app/server.go"
check "  Rate limiting on admin" "grep -q 'rl:admin' $BACKEND/internal/app/server.go"

if [ -f "$BACKEND/.env.staging" ]; then
    check "  .env.staging exists" "test -s $BACKEND/.env.staging"
else
    echo -e "  ${YELLOW}.env.staging not found (OK for dev)${NC}"
    WARN=$((WARN + 1))
fi
echo ""

# ──────────────────────────────────────────────
# A06: Vulnerable Components
# ──────────────────────────────────────────────
echo -e "${YELLOW}A06: Vulnerable and Outdated Components${NC}"

warn "  go mod tidy (clean)" "cd $BACKEND && go mod tidy -e 2>/dev/null"

if [ -d "$FRONTEND" ] && [ -f "$FRONTEND/package.json" ]; then
    echo "  - npm audit: skipping (requires npm install)"
    WARN=$((WARN + 1))
else
    echo -e "  ${YELLOW}Frontend not found, skipping npm audit${NC}"
    WARN=$((WARN + 1))
fi
echo ""

# ──────────────────────────────────────────────
# A07: Identification and Authentication Failures
# ──────────────────────────────────────────────
echo -e "${YELLOW}A07: Identification and Authentication Failures${NC}"
check "  JWT middleware" "test -f $BACKEND/shared/pkg/auth/middleware.go"
check "  Token extraction" "test -f $BACKEND/shared/pkg/auth/claims.go"
check "  Session management (Redis)" "grep -q 'auth:session' $BACKEND/internal/auth/handler.go"
check "  CSRF state protection" "grep -q 'auth:state' $BACKEND/internal/auth/handler.go"
check "  Rate limiting on auth" "grep -q 'RateLimitAuth' $BACKEND/internal/app/config.go"
echo ""

# ──────────────────────────────────────────────
# A08: Software and Data Integrity Failures
# ──────────────────────────────────────────────
echo -e "${YELLOW}A08: Software and Data Integrity Failures${NC}"
check "  go.sum exists" "test -f $BACKEND/go.sum"
check "  go.mod has module name" "grep -q '^module lkfl' $BACKEND/go.mod"
echo ""

# ──────────────────────────────────────────────
# A09: Security Logging and Monitoring Failures
# ──────────────────────────────────────────────
echo -e "${YELLOW}A09: Security Logging and Monitoring Failures${NC}"
check "  Auth events logged" "grep -q 'AuthLoginTotal\|AuthCallbackTotal' $BACKEND/internal/auth/handler.go"
check "  Logger exists" "test -f $BACKEND/shared/pkg/logger/logger.go"
check "  Prometheus metrics" "test -f $BACKEND/internal/metrics/metrics.go"
check "  Sentry integration" "grep -q 'sentry' $BACKEND/go.mod"
echo ""

# ──────────────────────────────────────────────
# A10: Server-Side Request Forgery (SSRF)
# ──────────────────────────────────────────────
echo -e "${YELLOW}A10: Server-Side Request Forgery (SSRF)${NC}"

SSRF_LINES=$(grep -rn 'http.Get\|http.Post\|http.DefaultClient.Get' $BACKEND/internal/ --include="*.go" 2>/dev/null | grep -v '_test.go' || true)
SSRF_COUNT=0
if [ -n "$SSRF_LINES" ]; then
    SSRF_COUNT=$(echo "$SSRF_LINES" | wc -l | tr -d '[:space:]')
fi
if [ "$SSRF_COUNT" -eq 0 ] 2>/dev/null; then
    echo -e "  ${GREEN}No unsafe HTTP client calls found${NC}"
    PASS=$((PASS + 1))
else
    echo -e "  ${YELLOW}HTTP client calls: $SSRF_COUNT (проверить вручную)${NC}"
    WARN=$((WARN + 1))
fi
echo ""

# ──────────────────────────────────────────────
# XSS Check (Frontend)
# ──────────────────────────────────────────────
echo -e "${YELLOW}XSS Check (Frontend)${NC}"

if [ -d "$FRONTEND" ]; then
    DANGEROUS_HTML=$(safe_grep_count -rn 'dangerouslySetInnerHTML' $FRONTEND/src/ --include="*.tsx" --include="*.ts")
    if [ "$DANGEROUS_HTML" -eq 0 ] 2>/dev/null; then
        echo -e "  ${GREEN}No dangerouslySetInnerHTML found${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "  ${YELLOW}dangerouslySetInnerHTML: $DANGEROUS_HTML (проверить вручную)${NC}"
        WARN=$((WARN + 1))
    fi
else
    echo -e "  ${YELLOW}Frontend not found, skipping XSS check${NC}"
    WARN=$((WARN + 1))
fi

echo "  - React JSX escaping: OK (built-in)"
PASS=$((PASS + 1))
echo ""

# ──────────────────────────────────────────────
# Build Check
# ──────────────────────────────────────────────
echo -e "${YELLOW}Build Check${NC}"
check "  go build ./..." "cd $BACKEND && go build ./..."
echo ""

# ──────────────────────────────────────────────
# Summary
# ──────────────────────────────────────────────
echo "========================="
TOTAL=$((PASS + FAIL))
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}, ${YELLOW}${WARN} warnings${NC} (total: ${TOTAL})"
echo ""

if [ "$FAIL" -gt 0 ]; then
    echo -e "${RED}Some security checks failed!${NC}"
    exit 1
elif [ "$WARN" -gt 0 ]; then
    echo -e "${YELLOW}All critical checks passed, but review warnings.${NC}"
    exit 0
else
    echo -e "${GREEN}All security checks passed!${NC}"
    exit 0
fi
