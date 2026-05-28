#!/bin/bash
# ============================================================================
# LKFL — Cosign Setup Script
# Настройка подписи Docker-образов с помощью Sigstore Cosign.
#
# Использование:
#   ./docker/cosign-setup.sh init          # Генерация ключевой пары
#   ./docker/cosign-setup.sh sign IMAGE    # Подпись образа
#   ./docker/cosign-setup.sh verify IMAGE  # Верификация подписи
#   ./docker/cosign-setup.sh key-info      # Информация о ключах
# ============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_DIR="${SCRIPT_DIR}/keys"
COSIGN_KEY="${KEYS_DIR}/cosign.key"
COSIGN_CRT="${KEYS_DIR}/cosign.crt"
COSIGN_PASSWORD_FILE="${KEYS_DIR}/.cosign-password"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# ============================================================================
# Проверка наличия cosign
# ============================================================================
check_cosign() {
    if ! command -v cosign &> /dev/null; then
        log_error "cosign не установлен."
        echo ""
        echo "Установка:"
        echo "  # macOS:  brew install sigstore/tap/cosign"
        echo "  # Linux:  curl -sSfL https://raw.githubusercontent.com/sigstore/cosign/main/install.sh | sh"
        echo "  # Debian: wget https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64 -O /usr/local/bin/cosign && chmod +x /usr/local/bin/cosign"
        exit 1
    fi
    log_info "cosign version: $(cosign version 2>/dev/null || echo 'unknown')"
}

# ============================================================================
# Init — генерация ключевой пары cosign
# ============================================================================
cmd_init() {
    check_cosign

    mkdir -p "${KEYS_DIR}"

    if [[ -f "${COSIGN_KEY}" && -f "${COSIGN_CRT}" ]]; then
        log_warn "Ключи уже существуют в ${KEYS_DIR}/"
        log_warn "Для перегенерации удалите: rm -rf ${KEYS_DIR}"
        exit 0
    fi

    # Генерация пароля для ключа
    local password
    password=$(openssl rand -base64 32)
    echo "${password}" > "${COSIGN_PASSWORD_FILE}"
    chmod 600 "${COSIGN_PASSWORD_FILE}"

    log_info "Генерация ключевой пары cosign..."

    # Генерация ключа с паролем
    COSIGN_PASSWORD="${password}" cosign generate-key-pair \
        --key "${COSIGN_KEY}" \
        --cert "${COSIGN_CRT}"

    chmod 600 "${COSIGN_KEY}" "${COSIGN_CRT}"

    log_info "Ключевая пара сгенерирована:"
    echo "  Key:  ${COSIGN_KEY}"
    echo "  Cert: ${COSIGN_CRT}"
    echo "  Password file: ${COSIGN_PASSWORD_FILE}"
    echo ""
    log_warn "Сохраните ${COSIGN_PASSWORD_FILE} в безопасном месте!"
    log_warn "Без пароля подпись и верификация невозможны."
    echo ""
    log_info "Для production используйте Kubernetes Secret или HashiCorp Vault:"
    echo "  kubectl create secret generic cosign-key \\"
    echo "    --from-file=key=${COSIGN_KEY} \\"
    echo "    --from-file=cert=${COSIGN_CRT} \\"
    echo "    --from-file=password=${COSIGN_PASSWORD_FILE}"
}

# ============================================================================
# Sign — подпись Docker-образа
# ============================================================================
cmd_sign() {
    local image="${1:?Использование: $0 sign <image>:<tag>}"
    check_cosign

    if [[ ! -f "${COSIGN_KEY}" ]]; then
        log_error "Ключи не найдены. Запустите: $0 init"
        exit 1
    fi

    log_info "Подпись образа: ${image}"

    COSIGN_PASSWORD=$(cat "${COSIGN_PASSWORD_FILE}") \
    cosign sign \
        --key "${COSIGN_KEY}" \
        --yes \
        "${image}"

    log_info "Образ ${image} подписан успешно."
}

# ============================================================================
# Verify — верификация подписи образа
# ============================================================================
cmd_verify() {
    local image="${1:?Использование: $0 verify <image>:<tag>}"
    check_cosign

    if [[ ! -f "${COSIGN_CRT}" ]]; then
        log_error "Сертификат не найден. Запустите: $0 init"
        exit 1
    fi

    log_info "Верификация образа: ${image}"

    cosign verify \
        --certificate "${COSIGN_CRT}" \
        "${image}"

    log_info "Подпись образа ${image} верифицирована успешно."
}

# ============================================================================
# Key Info — информация о ключах
# ============================================================================
cmd_key_info() {
    if [[ ! -f "${COSIGN_KEY}" || ! -f "${COSIGN_CRT}" ]]; then
        log_error "Ключи не найдены в ${KEYS_DIR}/"
        echo "Запустите: $0 init"
        exit 1
    fi

    log_info "Информация о ключах cosign:"
    echo ""
    echo "  Key file:  ${COSIGN_KEY}"
    echo "  Cert file: ${COSIGN_CRT}"
    echo "  Password:  ${COSIGN_PASSWORD_FILE}"
    echo ""

    # Показываем metadata сертификата
    if command -v openssl &> /dev/null; then
        echo "  Certificate details:"
        openssl x509 -in "${COSIGN_CRT}" -noout -subject -dates 2>/dev/null || true
    fi
}

# ============================================================================
# Attach — прикрепление подписи с claims (production)
# ============================================================================
cmd_attach() {
    local image="${1:?Использование: $0 attach <image>:<tag>}"
    check_cosign

    if [[ ! -f "${COSIGN_KEY}" ]]; then
        log_error "Ключи не найдены. Запустите: $0 init"
        exit 1
    fi

    log_info "Прикрепление подписи с claims для: ${image}"

    COSIGN_PASSWORD=$(cat "${COSIGN_PASSWORD_FILE}") \
    cosign sign \
        --key "${COSIGN_KEY}" \
        --yes \
        --attach \
        "${image}"

    log_info "Подпись с claims прикреплена для ${image}."
}

# ============================================================================
# Public Key — экспорт публичного ключа для верификаторов
# ============================================================================
cmd_public_key() {
    if [[ ! -f "${COSIGN_CRT}" ]]; then
        log_error "Сертификат не найден. Запустите: $0 init"
        exit 1
    fi

    log_info "Публичный ключ (certificate) для верификации:"
    echo ""
    cat "${COSIGN_CRT}"
    echo ""
    log_info "Сохраните в файл и распространите среди верификаторов."
}

# ============================================================================
# Main
# ============================================================================
main() {
    local command="${1:-help}"
    shift 2>/dev/null || true

    case "${command}" in
        init)
            cmd_init
            ;;
        sign)
            cmd_sign "$@"
            ;;
        verify)
            cmd_verify "$@"
            ;;
        key-info|info)
            cmd_key_info
            ;;
        attach)
            cmd_attach "$@"
            ;;
        public-key|pubkey)
            cmd_public_key
            ;;
        help|--help|-h|*)
            echo "LKFL Cosign Setup"
            echo ""
            echo "Использование: $0 <command> [args]"
            echo ""
            echo "Команды:"
            echo "  init              Генерация ключевой пары cosign"
            echo "  sign <image>      Подпись Docker-образа"
            echo "  verify <image>    Верификация подписи образа"
            echo "  key-info          Информация о ключах"
            echo "  attach <image>    Прикрепление подписи с claims"
            echo "  public-key        Экспорт публичного ключа"
            echo "  help              Показать помощь"
            echo ""
            echo "Примеры:"
            echo "  $0 init"
            echo "  $0 sign lkfl-server:v1.0.0"
            echo "  $0 verify lkfl-server:v1.0.0"
            ;;
    esac
}

main "$@"
