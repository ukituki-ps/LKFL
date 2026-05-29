#!/usr/bin/env bash
# LKFL — Post-deploy smoke test
# Проверяет базовую работоспособность после деплоя.
#
# Запуск:
#   ./infra/smoke-test.sh                          # по умолчанию staging
#   ./infra/smoke-test.sh https://dev.april.ukituki.tech
#
# Выход: 0 = все ок, 1 = есть проблемы

set -euo pipefail

BASE_URL="${1:-https://dev.april.ukituki.tech}"
FAILED=0
PASSED=0

pass() {
    echo "  ✅ $1"
    PASSED=$((PASSED + 1))
}

fail() {
    echo "  ❌ $1"
    FAILED=$((FAILED + 1))
}

echo "======================================"
echo "Smoke test: $BASE_URL"
echo "======================================"
echo ""

# ─── 1. Health check ───
echo "1. Health check"
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/healthz" 2>/dev/null || echo "000")
if [[ "$HEALTH" == "200" ]]; then
    pass "GET /healthz → 200"
else
    fail "GET /healthz → $HEALTH (expected 200)"
fi
echo ""

# ─── 2. Nginx health ───
echo "2. Nginx health"
NGINX_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/nginx-health" 2>/dev/null || echo "000")
if [[ "$NGINX_HEALTH" == "200" ]]; then
    pass "GET /nginx-health → 200"
else
    fail "GET /nginx-health → $NGINX_HEALTH (expected 200)"
fi
echo ""

# ─── 3. Login redirect URL (КРИТИЧНО) ───
echo "3. Login redirect (KEYCLOAK_PUBLIC_URL)"
# Следующий redirect с GET /api/v1/auth/login и проверяем Location header
LOGIN_REDIRECT=$(curl -s -o /dev/null -w "%{redirect_url}" -L --max-redirs 0 "$BASE_URL/api/v1/auth/login" 2>/dev/null || echo "")
if [[ -n "$LOGIN_REDIRECT" ]]; then
    # Проверка: URL не должен содержать internal Docker hostname
    if echo "$LOGIN_REDIRECT" | grep -q "keycloak:8080"; then
        fail "Login redirect содержит internal hostname 'keycloak:8080': $LOGIN_REDIRECT"
        fail "KEYCLOAK_PUBLIC_URL не настроен или равен KEYCLOAK_ISSUER"
    elif echo "$LOGIN_REDIRECT" | grep -q "localhost"; then
        fail "Login redirect содержит 'localhost': $LOGIN_REDIRECT"
    else
        pass "Login redirect URL не содержит internal hostname"
        echo "       → $LOGIN_REDIRECT"
    fi
else
    # curl не получил redirect — может быть 500 или другой код
    LOGIN_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/auth/login" 2>/dev/null || echo "000")
    if [[ "$LOGIN_CODE" == "302" ]]; then
        # Есть redirect но curl не показал URL — пробуем другой способ
        LOGIN_LOCATION=$(curl -s -I "$BASE_URL/api/v1/auth/login" 2>/dev/null | grep -i "^location:" | awk '{print $2}' | tr -d '\r')
        if echo "$LOGIN_LOCATION" | grep -q "keycloak:8080"; then
            fail "Login redirect содержит internal hostname: $LOGIN_LOCATION"
        else
            pass "Login redirect OK: $LOGIN_LOCATION"
        fi
    else
        fail "GET /api/v1/auth/login → $LOGIN_CODE (expected 302)"
    fi
fi
echo ""

# ─── 4. Keycloak accessibility ───
echo "4. Keycloak через Nginx proxy"
# Discovery endpoint конкретного realm (Keycloak не поддерживает /realms listing)
KC_DISCOVERY=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/realms/lkfl-sdek/.well-known/openid-configuration" 2>/dev/null || echo "000")
if [[ "$KC_DISCOVERY" == "200" ]]; then
    # Проверим что issuer содержит public URL, не internal hostname
    KC_ISSUER=$(curl -s "$BASE_URL/realms/lkfl-sdek/.well-known/openid-configuration" 2>/dev/null | grep -o '"issuer":"[^"]*"' | head -1)
    if echo "$KC_ISSUER" | grep -q "keycloak:8080"; then
        fail "Keycloak issuer содержит internal hostname: $KC_ISSUER"
    else
        pass "Keycloak discovery → 200, issuer OK: $KC_ISSUER"
    fi
else
    fail "Keycloak discovery → $KC_DISCOVERY (Keycloak недоступен через Nginx proxy)"
fi
echo ""

# ─── 5. Frontend loads ───
echo "5. Frontend SPA"
FRONTEND=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/" 2>/dev/null || echo "000")
if [[ "$FRONTEND" == "200" ]]; then
    pass "GET / → 200 (frontend загружается)"
else
    fail "GET / → $FRONTEND (expected 200)"
fi
echo ""

# ─── 6. API endpoint ───
echo "6. API endpoint"
API=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/engagements/" 2>/dev/null || echo "000")
# Ожидаем 401 (unauthorized) — значит API работает, просто нет токена
if [[ "$API" == "401" ]]; then
    pass "GET /api/v1/engagements/ → 401 (API работает, auth required)"
elif [[ "$API" == "404" ]]; then
    fail "GET /api/v1/engagements/ → 404 (API route не найден)"
else
    fail "GET /api/v1/engagements/ → $API (expected 401)"
fi
echo ""

# ─── Итог ───
echo "======================================"
echo "Result: $PASSED passed, $FAILED failed"
echo "======================================"

if [[ $FAILED -gt 0 ]]; then
    exit 1
fi
exit 0
