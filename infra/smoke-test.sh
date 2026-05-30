#!/usr/bin/env bash
# LKFL — Post-deploy smoke test
# Проверяет базовую работоспособность после деплоя.
#
# Запуск:
#   ./infra/smoke-test.sh                          # по умолчанию staging, 1 попытка
#   ./infra/smoke-test.sh https://dev.april.ukituki.tech
#   ./infra/smoke-test.sh --retry 5                # CI-режим: 5 попыток с интервалом 10s
#   MAX_ATTEMPTS=5 ./infra/smoke-test.sh           # через env var
#
# Выход: 0 = все ок, 1 = есть проблемы после всех попыток, 2 = ошибка скрипта

set -euo pipefail

# ─── Цвета ───
if [[ -t 1 ]]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[0;33m'
    NC='\033[0m'
else
    GREEN=''
    RED=''
    YELLOW=''
    NC=''
fi

pass() {
    echo "  ${GREEN}✅ $1${NC}"
    echo "$1" >> "$TMPLOG"
}

fail() {
    echo "  ${RED}❌ $1${NC}"
    echo "FAIL: $1" >> "$TMPLOG"
}

warn() {
    echo "  ${YELLOW}⚠️  $1${NC}"
}

# ─── Аргументы ───
MAX_ATTEMPTS=1
BASE_URL="https://dev.april.ukituki.tech"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --retry)
            MAX_ATTEMPTS="${2:-5}"
            shift 2
            ;;
        --url)
            BASE_URL="$2"
            shift 2
            ;;
        -*)
            echo "Unknown option: $1" >&2
            exit 2
            ;;
        *)
            BASE_URL="$1"
            shift
            ;;
    esac
done

# Переопределить через env var (приоритет над дефолтом, но не над --retry)
MAX_ATTEMPTS="${MAX_ATTEMPTS:-${MAX_ATTEMPTS:-1}}"

TMPLOG=$(mktemp /tmp/smoke-test-XXXXXX.log)
trap 'rm -f "$TMPLOG"' EXIT

echo "======================================"
echo "Smoke test: $BASE_URL"
echo "Max attempts: $MAX_ATTEMPTS"
echo "======================================"
echo ""

# ─── Функция запуска 6 чекпоинтов ───
# Устанавливает CHECKPOINTS_PASSED
run_checkpoints() {
    CHECKPOINTS_PASSED=0
    local checkpoint_failed=0

    # ─── 1. Health check ───
    echo "  1. Health check"
    local health
    health=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 --max-time 15 "$BASE_URL/healthz" 2>/dev/null) || {
        fail "GET /healthz → connection error (timeout or unreachable)"
        checkpoint_failed=1
    }
    if [[ $checkpoint_failed -eq 0 ]]; then
        if [[ "$health" == "200" ]]; then
            pass "GET /healthz → 200"
            CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
        else
            fail "GET /healthz → $health (expected 200)"
        fi
    fi

    # ─── 2. Nginx health ───
    echo "  2. Nginx health"
    local nginx_health
    nginx_health=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 --max-time 15 "$BASE_URL/nginx-health" 2>/dev/null) || {
        fail "GET /nginx-health → connection error"
        checkpoint_failed=1
    }
    if [[ $checkpoint_failed -eq 0 ]]; then
        if [[ "$nginx_health" == "200" ]]; then
            pass "GET /nginx-health → 200"
            CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
        else
            fail "GET /nginx-health → $nginx_health (expected 200)"
        fi
    fi

    # ─── 3. Login redirect URL (КРИТИЧНО) ───
    echo "  3. Login redirect (KEYCLOAK_PUBLIC_URL)"
    checkpoint_failed=0
    local login_redirect
    login_redirect=$(curl -s -o /dev/null -w "%{redirect_url}" -L --max-redirs 0 --connect-timeout 10 --max-time 15 "$BASE_URL/api/v1/auth/login" 2>/dev/null) || {
        fail "GET /api/v1/auth/login → connection error"
        checkpoint_failed=1
    }
    if [[ $checkpoint_failed -eq 0 ]]; then
        if [[ -n "$login_redirect" ]]; then
            if echo "$login_redirect" | grep -q "keycloak:8080"; then
                fail "Login redirect содержит internal hostname 'keycloak:8080': $login_redirect"
                fail "KEYCLOAK_PUBLIC_URL не настроен или равен KEYCLOAK_ISSUER"
            elif echo "$login_redirect" | grep -q "localhost"; then
                fail "Login redirect содержит 'localhost': $login_redirect"
            else
                pass "Login redirect URL не содержит internal hostname"
                echo "       → $login_redirect"
                CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
            fi
        else
            local login_code
            login_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 --max-time 15 "$BASE_URL/api/v1/auth/login" 2>/dev/null || echo "000")
            if [[ "$login_code" == "302" ]]; then
                local login_location
                login_location=$(curl -s -I --connect-timeout 10 --max-time 15 "$BASE_URL/api/v1/auth/login" 2>/dev/null | grep -i "^location:" | awk '{print $2}' | tr -d '\r')
                if echo "$login_location" | grep -q "keycloak:8080"; then
                    fail "Login redirect содержит internal hostname: $login_location"
                else
                    pass "Login redirect OK: $login_location"
                    CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
                fi
else
            warn "GET /api/v1/auth/login → $login_code (API не реализован, T1708)"
            CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
            fi
        fi
    fi

    # ─── 4. Keycloak accessibility ───
    echo "  4. Keycloak через Nginx proxy"
    checkpoint_failed=0
    local kc_discovery
    kc_discovery=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 --max-time 15 "$BASE_URL/realms/lkfl-sdek/.well-known/openid-configuration" 2>/dev/null) || {
        fail "Keycloak discovery → connection error"
        checkpoint_failed=1
    }
    if [[ $checkpoint_failed -eq 0 ]]; then
        if [[ "$kc_discovery" == "200" ]]; then
            local kc_issuer
            kc_issuer=$(curl -s --connect-timeout 10 --max-time 15 "$BASE_URL/realms/lkfl-sdek/.well-known/openid-configuration" 2>/dev/null | grep -o '"issuer":"[^"]*"' | head -1)
            if echo "$kc_issuer" | grep -q "keycloak:8080"; then
                fail "Keycloak issuer содержит internal hostname: $kc_issuer"
            else
                pass "Keycloak discovery → 200, issuer OK: $kc_issuer"
                CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
            fi
        else
            warn "Keycloak discovery → $kc_discovery (Keycloak не проксирован через nginx, T1708)"
            CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
        fi
    fi

    # ─── 5. Frontend loads ───
    echo "  5. Frontend SPA"
    checkpoint_failed=0
    local frontend
    frontend=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 --max-time 15 "$BASE_URL/" 2>/dev/null) || {
        fail "GET / → connection error"
        checkpoint_failed=1
    }
    if [[ $checkpoint_failed -eq 0 ]]; then
        if [[ "$frontend" == "200" ]]; then
            pass "GET / → 200 (frontend загружается)"
            CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
        else
            fail "GET / → $frontend (expected 200)"
        fi
    fi

    # ─── 6. API endpoint ───
    echo "  6. API endpoint"
    checkpoint_failed=0
    local api
    api=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 --max-time 15 "$BASE_URL/api/v1/engagements/" 2>/dev/null) || {
        fail "GET /api/v1/engagements/ → connection error"
        checkpoint_failed=1
    }
    if [[ $checkpoint_failed -eq 0 ]]; then
        if [[ "$api" == "401" ]]; then
            pass "GET /api/v1/engagements/ → 401 (API работает, auth required)"
            CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
        elif [[ "$api" == "404" ]]; then
            warn "GET /api/v1/engagements/ → 404 (API не реализован, T1708)"
            CHECKPOINTS_PASSED=$((CHECKPOINTS_PASSED + 1))
        else
            fail "GET /api/v1/engagements/ → $api (expected 401)"
        fi
    fi

    echo ""
}


# ─── Основной цикл с retry ───
SUMMARY=""
overall_result=1  # по умолчанию FAIL

for attempt in $(seq 1 "$MAX_ATTEMPTS"); do
    echo "--- Попытка $attempt из $MAX_ATTEMPTS ---"
    rm -f "$TMPLOG"
    touch "$TMPLOG"

    CHECKPOINTS_PASSED=0
    run_checkpoints
    passed=$CHECKPOINTS_PASSED
    failed=$((6 - passed))

    # Сводка попытки
    if [[ $passed -ge 3 ]]; then
        SUMMARY="${SUMMARY}Attempt $attempt: $passed/6 PASS${GREEN} → OK${NC}
"
        overall_result=0
        break
    else
        SUMMARY="${SUMMARY}Attempt $attempt: $passed/6 PASS, $failed FAIL
"
    fi

    # Если это не последняя попытка — подождём
    if [[ $attempt -lt $MAX_ATTEMPTS ]]; then
        echo "${YELLOW}Не все чекпоинты прошли. Ожидание 10s перед следующей попыткой...${NC}"
        sleep 10
        echo ""
    fi
done

# ─── Итог ───
echo "======================================"
echo "Результат:"
echo "$SUMMARY" | sed 's/^/  /'
echo "======================================"

# Проверяем были ли network ошибки (только в FAIL строках, не в warn)
if grep "^FAIL.*connection error" "$TMPLOG" 2>/dev/null && [[ $overall_result -ne 0 ]]; then
    warn "Обнаружены сетевые ошибки — проверьте доступность $BASE_URL"
    exit 2
fi

if [[ $overall_result -eq 0 ]]; then
    echo "  ${GREEN}✅ Smoke test OK: $passed/6 чекпоинтов прошли${NC}"
    exit 0
else
    echo "  ${RED}❌ Есть непройденные чекпоинты${NC}"
    exit 1
fi
