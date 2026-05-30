#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# setup-all.sh — устанавливает nginx на serverAi + serverPr01
# ============================================================================
#
# Назначение:
#   Мастер-скрипт для установки nginx конфигов на оба сервера.
#   Выполняется с МАШИНЫ РАЗРАБОТЧИКА (не на сервере).
#
# Требования:
#   - SSH key настроен (ssh serverAi, ssh serverPr01 в ~/.ssh/config)
#   - Репозиторий LKFL склонирован на машину разработчика
#
# Использование:
#   bash infra/scripts/setup-all.sh
#
# ============================================================================

# --- Хосты из ~/.ssh/config ---
SERVER_AI="serverAi"        # ~/.ssh/config host → 192.168.1.27
SERVER_PR01="serverPr01"    # ~/.ssh/config host → 192.168.1.46

# --- Пути ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
NGINX_DIR="${SCRIPT_DIR}/../nginx"

REMOTE_SCRIPT_DIR="/tmp/lkfl-setup"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }
log_step()  { echo -e "${CYAN}[STEP]${NC} $*"; }

# --- Проверка существования конфигов ---
if [ ! -f "${NGINX_DIR}/serverAi.conf" ]; then
    log_error "Конфиг serverAi.conf не найден в ${NGINX_DIR}"
    exit 1
fi

if [ ! -f "${NGINX_DIR}/serverPr01-internal.conf" ]; then
    log_error "Конфиг serverPr01-internal.conf не найден в ${NGINX_DIR}"
    exit 1
fi

if [ ! -f "${SCRIPT_DIR}/setup-nginx-serverAi.sh" ]; then
    log_error "Скрипт setup-nginx-serverAi.sh не найден"
    exit 1
fi

if [ ! -f "${SCRIPT_DIR}/setup-nginx-serverPr01.sh" ]; then
    log_error "Скрипт setup-nginx-serverPr01.sh не найден"
    exit 1
fi

log_info "=============================================="
log_info "  LKFL — Установка nginx на сервера"
log_info "=============================================="
log_info "Проект: ${PROJECT_DIR}"
log_info "serverAi:  ${SERVER_AI} (192.168.1.27)"
log_info "serverPr01: ${SERVER_PR01} (192.168.1.46)"
log_info ""

# ==========================================================================
# serverAi (192.168.1.27)
# ==========================================================================
log_step "===== Установка на serverAi ====="

# Копируем конфиг и скрипт на serverAi
log_info "Копирование файлов на ${SERVER_AI}..."
ssh "${SERVER_AI}" "mkdir -p ${REMOTE_SCRIPT_DIR}"
scp "${NGINX_DIR}/serverAi.conf" "${SERVER_AI}:${REMOTE_SCRIPT_DIR}/serverAi.conf"
scp "${SCRIPT_DIR}/setup-nginx-serverAi.sh" "${SERVER_AI}:${REMOTE_SCRIPT_DIR}/setup-nginx-serverAi.sh"
scp "${NGINX_DIR}/serverPr01-internal.conf" "${SERVER_AI}:${REMOTE_SCRIPT_DIR}/serverPr01-internal.conf"

# Создаём директорию nginx на удалённом сервере для структуры
ssh "${SERVER_AI}" "mkdir -p ${REMOTE_SCRIPT_DIR}/nginx"
scp "${NGINX_DIR}/serverAi.conf" "${SERVER_AI}:${REMOTE_SCRIPT_DIR}/nginx/serverAi.conf"

# Запускаем скрипт установки
log_info "Запуск установки на ${SERVER_AI}..."
if ssh "${SERVER_AI}" "bash ${REMOTE_SCRIPT_DIR}/setup-nginx-serverAi.sh"; then
    log_info "✅ serverAi: установка завершена успешно"
else
    log_error "❌ serverAi: установка не удалась"
    exit 1
fi

# Очищаем временные файлы
ssh "${SERVER_AI}" "rm -rf ${REMOTE_SCRIPT_DIR}"

# Проверяем health
log_info "Проверка health на serverAi..."
if ssh "${SERVER_AI}" "curl -sf http://127.0.0.1:18000/nginx-health"; then
    log_info "✅ serverAi: health check OK"
else
    log_warn "⚠️ serverAi: health check не прошёл"
fi

log_info ""

# ==========================================================================
# serverPr01 (192.168.1.46)
# ==========================================================================
log_step "===== Установка на serverPr01 ====="

# Копируем конфиг и скрипт на serverPr01
log_info "Копирование файлов на ${SERVER_PR01}..."
ssh "${SERVER_PR01}" "mkdir -p ${REMOTE_SCRIPT_DIR}"
scp "${NGINX_DIR}/serverPr01-internal.conf" "${SERVER_PR01}:${REMOTE_SCRIPT_DIR}/serverPr01-internal.conf"
scp "${SCRIPT_DIR}/setup-nginx-serverPr01.sh" "${SERVER_PR01}:${REMOTE_SCRIPT_DIR}/setup-nginx-serverPr01.sh"

# Создаём директорию nginx на удалённом сервере
ssh "${SERVER_PR01}" "mkdir -p ${REMOTE_SCRIPT_DIR}/nginx"
scp "${NGINX_DIR}/serverPr01-internal.conf" "${SERVER_PR01}:${REMOTE_SCRIPT_DIR}/nginx/serverPr01-internal.conf"

# Запускаем скрипт установки
log_info "Запуск установки на ${SERVER_PR01}..."
if ssh "${SERVER_PR01}" "bash ${REMOTE_SCRIPT_DIR}/setup-nginx-serverPr01.sh"; then
    log_info "✅ serverPr01: установка завершена успешно"
else
    log_error "❌ serverPr01: установка не удалась"
    exit 1
fi

# Очищаем временные файлы
ssh "${SERVER_PR01}" "rm -rf ${REMOTE_SCRIPT_DIR}"

# Проверяем health
log_info "Проверка health на serverPr01..."
if ssh "${SERVER_PR01}" "curl -sf http://127.0.0.1:18000/nginx-health"; then
    log_info "✅ serverPr01: health check OK"
else
    log_warn "⚠️ serverPr01: health check не прошёл"
fi

log_info ""
log_info "=============================================="
log_info "  Установка завершена!"
log_info "=============================================="
log_info "serverAi:  http://${SERVER_AI}:18000/nginx-health"
log_info "serverPr01: http://${SERVER_PR01}:18000/nginx-health"
log_info ""
log_info "Проверка end-to-end через serverPr01:"
log_info "  curl http://${SERVER_PR01}:18000/healthz"
