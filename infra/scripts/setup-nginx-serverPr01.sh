#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# setup-nginx-serverPr01.sh — устанавливает nginx на 18000 для LKFL на serverPr01
# ============================================================================
#
# Назначение:
#   Копирует конфиг serverPr01-internal.conf и активирует его в nginx.
#   serverPr01 работает как обратный прокси к serverAi:18000.
#
# Требования:
#   - Выполняется на serverPr01 (192.168.1.46) от root
#   - nginx установлен
#
# Использование:
#   bash setup-nginx-serverPr01.sh
#
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NGINX_DIR="${SCRIPT_DIR}/../nginx"

CONF_SOURCE="${NGINX_DIR}/serverPr01-internal.conf"
CONF_DEST="/etc/nginx/sites-available/lkfl-internal.conf"
LINK_DEST="/etc/nginx/sites-enabled/lkfl-internal.conf"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# --- Проверка root ---
if [ "$(id -u)" -ne 0 ]; then
    log_error "Скрипт должен выполняться от root"
    exit 1
fi

# --- Проверка исходного конфига ---
if [ ! -f "${CONF_SOURCE}" ]; then
    log_error "Конфиг не найден: ${CONF_SOURCE}"
    exit 1
fi

log_info "===== Установка nginx для LKFL на serverPr01 ====="
log_info "Источник конфига: ${CONF_SOURCE}"

# --- Копируем конфиг в sites-available ---
log_info "Копирование конфига в ${CONF_DEST}..."
cp "${CONF_SOURCE}" "${CONF_DEST}"
log_info "Конфиг скопирован."

# --- Создаём символическую ссылку в sites-enabled ---
log_info "Создание символической ссылки: ${LINK_DEST}..."
ln -sf "${CONF_DEST}" "${LINK_DEST}"
log_info "Ссылка создана."

# --- Тест конфигурации nginx ---
log_info "Тест конфигурации nginx..."
if nginx -t 2>&1; then
    log_info "Конфигурация валидна."
else
    log_error "Конфигурация nginx недействительна!"
    log_error "Откат: удаляем ссылку..."
    rm -f "${LINK_DEST}"
    exit 1
fi

# --- Перезагрузка nginx ---
log_info "Перезагрузка nginx..."
if nginx -s reload 2>&1; then
    log_info "nginx перезапущен успешно."
else
    log_warn "nginx не был запущен, пытаемся запустить..."
    nginx
    log_info "nginx запущен."
fi

log_info "===== Установка завершена успешно ====="
log_info "Nginx слушает порт 18000 и проксирует на serverAi:18000"
log_info "Проверка: curl http://127.0.0.1:18000/nginx-health"
