#!/usr/bin/env bash
set -euo pipefail

# Переключает тему Keycloak на бренд tenant'а.
# Usage: switch-tenant-theme.sh [tenant-slug]
#   default  → April teal (нейтральный)
#   sdek     → СДЭК красный

TENANT=${1:-default}
THEME_DIR="$(cd "$(dirname "$0")" && pwd)/theme/lkfl/login/resources/css"
PROPS_LOGIN="$(cd "$(dirname "$0")" && pwd)/theme/lkfl/login/theme.properties"
PROPS_ACCOUNT="$(cd "$(dirname "$0")" && pwd)/theme/lkfl/account/theme.properties"

# Validate CSS file exists
BRAND_CSS="brand-${TENANT}.css"
if [ ! -f "${THEME_DIR}/${BRAND_CSS}" ]; then
  echo "Error: ${THEME_DIR}/${BRAND_CSS} not found"
  exit 1
fi

# Update theme.properties
sed -i "s|^styles=.*|styles=css/${BRAND_CSS}, css/login.css, css/theme-toggle.css|" "${PROPS_LOGIN}"
sed -i "s|^styles=.*|styles=css/${BRAND_CSS}, css/login.css, css/theme-toggle.css|" "${PROPS_ACCOUNT}"

echo "Theme switched to '${TENANT}' for login + account console"
echo "  ${PROPS_LOGIN}"
echo "  ${PROPS_ACCOUNT}"
echo "  Restart Keycloak to apply"
