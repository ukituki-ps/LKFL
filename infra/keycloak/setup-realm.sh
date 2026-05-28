#!/bin/bash
# infra/keycloak/setup-realm.sh
# Создаёт realm «lkfl-sdek» через Keycloak Admin API
# (для production, где --import-realm недоступен)
#
# Использование:
#   KEYCLOAK_ADMIN_URL=http://localhost:8081 \
#   ADMIN_TOKEN=xxx \
#   bash infra/keycloak/setup-realm.sh
#
# Получить ADMIN_TOKEN:
#   curl -X POST http://localhost:8081/realms/master/protocol/openid-connect/token \
#     -d "grant_type=password" \
#     -d "client_id=admin-cli" \
#     -d "username=${KEYCLOAK_ADMIN}" \
#     -d "password=${KEYCLOAK_ADMIN_PASSWORD}" \
#   | jq -r '.access_token'

set -euo pipefail

REALM_FILE="infra/keycloak/realm-lkfl-sdek.json"

if [ ! -f "$REALM_FILE" ]; then
    echo "ERROR: Realm file not found: $REALM_FILE"
    exit 1
fi

if [ -z "${KEYCLOAK_ADMIN_URL:-}" ]; then
    echo "ERROR: KEYCLOAK_ADMIN_URL is not set"
    echo "  Example: export KEYCLOAK_ADMIN_URL=http://localhost:8081"
    exit 1
fi

if [ -z "${ADMIN_TOKEN:-}" ]; then
    echo "ERROR: ADMIN_TOKEN is not set"
    echo "  Get token from Keycloak master realm admin-cli client"
    exit 1
fi

echo "Creating realm 'lkfl-sdek' via Keycloak Admin API..."
echo "  URL: ${KEYCLOAK_ADMIN_URL}/admin/realms"
echo "  File: ${REALM_FILE}"
echo ""

RESPONSE=$(curl -sfX POST "${KEYCLOAK_ADMIN_URL}/admin/realms" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d "@${REALM_FILE}" \
  -w "\n%{http_code}")

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "201" ]; then
    echo "Realm 'lkfl-sdek' created successfully."
elif [ "$HTTP_CODE" = "409" ]; then
    echo "Realm 'lkfl-sdek' already exists."
    echo "Delete it first if you want to reimport:"
    echo "  curl -sfX DELETE '${KEYCLOAK_ADMIN_URL}/admin/realms/lkfl-sdek' \\"
    echo "    -H 'Authorization: Bearer ${ADMIN_TOKEN}'"
    exit 0
else
    echo "ERROR: Failed to create realm. HTTP ${HTTP_CODE}"
    echo "$BODY" | head -50
    exit 1
fi

echo ""
echo "Next steps:"
echo "  1. Verify realm: curl -sf '${KEYCLOAK_ADMIN_URL}/realms/lkfl-sdek/.well-known/openid-configuration'"
echo "  2. Get client secret for lkfl-service:"
echo "     curl -sf '${KEYCLOAK_ADMIN_URL}/admin/realms/lkfl-sdek/clients' \\"
echo "       -H 'Authorization: Bearer ${ADMIN_TOKEN}' \\"
echo "       | jq '.[] | select(.clientId==\"lkfl-service\") | .clientSecret'"
