#!/bin/bash
# T2205 — Запуск нагрузочных тестов k6
#
# Использует k6 для запуска всех сценариев нагрузки.
# Результаты сохраняются в JSON формате для CI artifacts.
#
# Требования:
#   - k6 v0.50+ (https://k6.io/docs/get-started/installation/)
#   - lkfl-server запущен на localhost:8080
#
# Использование:
#   ./loadtest/run.sh                          # все тесты
#   ./loadtest/run.sh catalog                  # только каталог
#   ./loadtest/run.sh auth                     # только auth
#   ./loadtest/run.sh profile                  # только profile
#   ./loadtest/run.sh combined                 # только combined
#   BASE_URL=http://staging.lkfl.local ./loadtest/run.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_DIR="${SCRIPT_DIR}/results"
BASE_URL="${BASE_URL:-http://localhost:8080}"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
EXIT_CODE=0

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# --- Утилиты ---

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# --- Проверка окружения ---

check_k6() {
    if ! command -v k6 &> /dev/null; then
        log_error "k6 не найден. Установите: https://k6.io/docs/get-started/installation/"
        exit 1
    fi

    local version
    version=$(k6 version 2>/dev/null | head -1)
    log_info "k6 версия: ${version}"
}

check_server() {
    log_info "Проверка доступности сервера: ${BASE_URL}"
    local http_code
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/healthz" 2>/dev/null || echo "000")

    if [ "${http_code}" != "200" ]; then
        log_error "Сервер недоступен (${BASE_URL}/api/healthz → ${http_code})"
        log_info "Запустите lkfl-server перед выполнением нагрузочных тестов"
        exit 1
    fi
    log_info "Сервер доступен (healthz: ${http_code})"
}

# --- Запуск теста ---

run_test() {
    local script="$1"
    local name="$2"

    log_info "═══════════════════════════════════════"
    log_info "Запуск теста: ${name}"
    log_info "Скрипт: ${script}"
    log_info "Цель: ${BASE_URL}"
    log_info "═══════════════════════════════════════"

    local output_file="${RESULTS_DIR}/${name}_${TIMESTAMP}.json"

    BASE_URL="${BASE_URL}" k6 run \
        --out json="${output_file}" \
        "${SCRIPT_DIR}/${script}" \
        2>&1

    local k6_exit=$?

    if [ ${k6_exit} -eq 0 ]; then
        log_info "✅ ${name}: ВСЕ THRESHOLDS ПРОЙДЕНЫ"
    else
        log_error "❌ ${name}: THRESHOLD(S) НЕ ПРОЙДЕНЫ"
        EXIT_CODE=1
    fi

    log_info "Результаты: ${output_file}"
    echo ""
}

# --- HTML отчёт ---

generate_html_report() {
    log_info "Генерация HTML отчёта..."

    local html_file="${RESULTS_DIR}/index.html"

    # Собираем все JSON файлы результатов
    local json_files=("${RESULTS_DIR}"/*_"${TIMESTAMP}".json)

    if [ ${#json_files[@]} -eq 0 ]; then
        log_warn "Нет JSON файлов для отчёта"
        return
    fi

    # Генерируем HTML отчёт из k6 JSON данных
    cat > "${html_file}" << 'HTMLHEADER'
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LKFL — Отчёт нагрузочного тестирования</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #333; margin-bottom: 10px; }
        .meta { color: #666; margin-bottom: 30px; font-size: 14px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(350px, 1fr)); gap: 20px; }
        .card { background: white; border-radius: 8px; padding: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .card h2 { color: #333; margin-bottom: 15px; font-size: 18px; border-bottom: 2px solid #eee; padding-bottom: 10px; }
        .metric { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #f0f0f0; }
        .metric:last-child { border-bottom: none; }
        .metric-name { color: #555; }
        .metric-value { font-weight: 600; }
        .pass { color: #22c55e; }
        .fail { color: #ef4444; }
        .summary { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .summary-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; }
        .summary-item { text-align: center; }
        .summary-value { font-size: 28px; font-weight: 700; }
        .summary-label { color: #666; font-size: 12px; margin-top: 5px; }
        table { width: 100%; border-collapse: collapse; margin-top: 10px; }
        th, td { padding: 8px 12px; text-align: left; border-bottom: 1px solid #eee; font-size: 13px; }
        th { background: #f8f8f8; font-weight: 600; color: #555; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔬 LKFL — Отчёт нагрузочного тестирования</h1>
        <div class="meta">
            Дата: <span id="testDate"></span> |
            Сервер: <span id="testServer"></span> |
            Тестов: <span id="testCount"></span>
        </div>
        <div class="summary" id="summary"></div>
        <div class="grid" id="results"></div>
    </div>
    <script>
HTMLHEADER

    cat >> "${html_file}" << 'HTMLSCRIPT'
        document.getElementById('testDate').textContent = new Date().toLocaleString('ru-RU');
        document.getElementById('testServer').textContent = '"${BASE_URL}"';

        const testData = $(cat "${RESULTS_DIR}"/*_${TIMESTAMP}.json 2>/dev/null | head -c 500000);

        // Парсим k6 JSON output (newline-delimited JSON)
        const lines = testData.trim().split('\n').filter(l => l.trim());
        const metrics = {};
        const rootGroups = {};
        let totalChecks = 0;
        let passedChecks = 0;

        for (const line of lines) {
            try {
                const data = JSON.parse(line);
                if (data.type === 'metric' && data.data && data.data.type === 'threshold') {
                    const name = data.data.name;
                    if (!metrics[name]) metrics[name] = {};
                    for (const [threshold, result] of Object.entries(data.data.thresholds || {})) {
                        metrics[name][threshold] = result;
                    }
                }
                if (data.type === 'metric' && data.data && data.data.type === 'gauge') {
                    if (data.data.name === 'http_reqs') {
                        rootGroups['http_reqs'] = data.data.value;
                    }
                }
                if (data.type === 'point' && data.metric === 'http_req_duration') {
                    if (!rootGroups['durations']) rootGroups['durations'] = [];
                    rootGroups['durations'].push(data.value);
                }
            } catch(e) {}
        }

        // Рендер карточек
        const resultsDiv = document.getElementById('results');
        const summaryDiv = document.getElementById('summary');

        // Подсчёт pass/fail
        let passCount = 0;
        let failCount = 0;
        for (const [name, thresholds] of Object.entries(metrics)) {
            for (const [threshold, result] of Object.entries(thresholds)) {
                if (result.ok) passCount++;
                else failCount++;
            }
        }

        // Summary
        const totalDuration = rootGroups['durations'] ?
            rootGroups['durations'].reduce((a, b) => a + b, 0) / rootGroups['durations'].length : 0;

        summaryDiv.innerHTML = `
            <div class="summary-grid">
                <div class="summary-item">
                    <div class="summary-value ${failCount === 0 ? 'pass' : 'fail'}">${passCount}</div>
                    <div class="summary-label">Пройдено</div>
                </div>
                <div class="summary-item">
                    <div class="summary-value ${failCount > 0 ? 'fail' : 'pass'}">${failCount}</div>
                    <div class="summary-label">Не пройдено</div>
                </div>
                <div class="summary-item">
                    <div class="summary-value">${Math.round(totalDuration)}</div>
                    <div class="summary-label">Среднее время (ms)</div>
                </div>
                <div class="summary-item">
                    <div class="summary-value">${(passCount + failCount)}</div>
                    <div class="summary-label">Всего проверок</div>
                </div>
            </div>
        `;

        document.getElementById('testCount').textContent = Object.keys(metrics).length;

        // Карточки по метрикам
        for (const [name, thresholds] of Object.entries(metrics)) {
            const card = document.createElement('div');
            card.className = 'card';

            let html = `<h2>${name}</h2>`;
            html += '<table><tr><th>Threshold</th><th>Результат</th></tr>';

            for (const [threshold, result] of Object.entries(thresholds)) {
                const statusClass = result.ok ? 'pass' : 'fail';
                const statusText = result.ok ? '✅ OK' : '❌ FAIL';
                html += `<tr><td>${threshold}</td><td class="${statusClass}">${statusText}</td></tr>`;
            }
            html += '</table>';
            card.innerHTML = html;
            resultsDiv.appendChild(card);
        }
    </script>
</body>
</html>
HTMLSCRIPT

    log_info "HTML отчёт: ${html_file}"
    log_info "Откройте в браузере: file://${html_file}"
}

# --- Main ---

main() {
    local target="${1:-all}"

    log_info "LKFL Load Testing — ${TIMESTAMP}"
    log_info "Results dir: ${RESULTS_DIR}"
    log_info "Target: ${target}"
    echo ""

    check_k6
    check_server

    mkdir -p "${RESULTS_DIR}"

    case "${target}" in
        catalog)
            run_test "catalog.js" "catalog"
            ;;
        auth)
            run_test "auth.js" "auth"
            ;;
        profile)
            run_test "profile.js" "profile"
            ;;
        combined)
            run_test "combined.js" "combined"
            ;;
        all|"")
            run_test "catalog.js" "catalog"
            run_test "auth.js" "auth"
            run_test "profile.js" "profile"
            run_test "combined.js" "combined"
            ;;
        *)
            log_error "Неизвестный тест: ${target}"
            log_info "Доступные: catalog, auth, profile, combined, all"
            exit 1
            ;;
    esac

    echo ""
    log_info "═══════════════════════════════════════"
    log_info "Все тесты завершены"
    log_info "Результаты: ${RESULTS_DIR}/"
    log_info "═══════════════════════════════════════"

    if [ ${EXIT_CODE} -eq 0 ]; then
        log_info "✅ ВСЕ THRESHOLDS ПРОЙДЕНЫ"
    else
        log_error "❌ НЕКОТОРЫЕ THRESHOLDS НЕ ПРОЙДЕНЫ"
    fi

    # Генерируем HTML отчёт
    generate_html_report

    exit ${EXIT_CODE}
}

main "$@"
