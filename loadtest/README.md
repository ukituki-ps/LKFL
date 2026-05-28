# LKFL — Нагрузочное тестирование (k6)

Нагрузочные тесты для backend LKFL платформы. Валидация производительности при целевых нагрузках F1 (Рабочий каталог).

## Требования

- **k6** v0.50+ — [установка](https://k6.io/docs/get-started/installation/)
- **lkfl-server** запущен на `localhost:8080`
- **curl** (для проверки healthz в run.sh)

### Установка k6

```bash
# Linux (deb)
sudo apt-get install -y apt-transport-https gnupg
sudo mkdir -p /etc/apt/keyrings
curl -s https://k6.io/misc/gpg/key.gpg | sudo gpg --dearmor -o /etc/apt/keyrings/loadimpact-archive-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/loadimpact-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/loadimpact.list
sudo apt-get update
sudo apt-get install -y k6

# macOS
brew install k6

# Docker
docker pull grafana/k6
```

## Быстрый старт

```bash
# Запуск всех тестов
./loadtest/run.sh

# Отдельный тест
./loadtest/run.sh catalog
./loadtest/run.sh auth
./loadtest/run.sh profile
./loadtest/run.sh combined

# С кастомным URL
BASE_URL=http://staging.lkfl.local ./loadtest/run.sh
```

## Сценарии

### 1. Каталог (`catalog.js`)

Тестирует endpoints каталога энгэйджментов.

| Параметр | Значение |
|----------|----------|
| Peak нагрузка | 500 виртуальных пользователей |
| P95 SLA | < 200ms |
| P99 SLA | < 500ms |
| Ошибки | < 1% |
| Длительность | ~2.5 мин |

**Эндпоинты:**
- `GET /api/v1/engagements` — список с пагинацией (60%)
- `GET /api/v1/engagements/:id` — детали (20%)
- `GET /api/v1/engagements?type=&status=` — фильтрация (10%)
- `GET /api/v1/engagements?search=` — поиск (10%)

### 2. Auth (`auth.js`)

Тестирует аутентификационные endpoints.

| Параметр | Значение |
|----------|----------|
| Peak нагрузка | 100 виртуальных пользователей |
| P95 SLA | < 500ms |
| P99 SLA | < 1000ms |
| Ошибки | < 1% |
| Длительность | ~1.5 мин |

**Эндпоинты:**
- `GET /api/v1/user/me` — проверка сессии (80%)
- `GET /api/v1/auth/callback` — OIDC callback (15%)
- `POST /api/v1/auth/register` — регистрация (5%)

### 3. User Profile (`profile.js`)

Тестирует endpoints профиля пользователя.

| Параметр | Значение |
|----------|----------|
| Peak нагрузка | 200 виртуальных пользователей |
| P95 SLA | < 100ms |
| P99 SLA | < 200ms |
| Ошибки | < 1% |
| Длительность | ~1.5 мин |

**Эндпоинты:**
- `GET /api/v1/user/me` — чтение профиля (90%)
- `PUT /api/v1/user/me` — обновление контактов (10%)

### 4. Комбинированная нагрузка (`combined.js`)

Имитирует реальное распределение трафика по всем модулям.

| Параметр | Значение |
|----------|----------|
| Peak нагрузка | 600 виртуальных пользователей |
| Общий P95 | < 300ms |
| Общий P99 | < 800ms |
| Ошибки | < 1% |
| Длительность | ~5 мин |

**Распределение:**
- 40% — Каталог (engagements)
- 20% — User Profile (me)
- 20% — User Engagements
- 10% — Dashboard
- 5% — Auth (callback)
- 5% — Рекомендации

## Результаты

После запуска результаты сохраняются в `loadtest/results/`:

```
loadtest/results/
├── catalog_20260115_143000.json    # JSON метрики k6
├── auth_20260115_143200.json
├── profile_20260115_143300.json
├── combined_20260115_143400.json
└── index.html                       # HTML отчёт
```

### Формат JSON

Каждый JSON файл содержит newline-delimited JSON (NDJSON) с метриками k6.
Парсинг для CI/CD:

```bash
# Извлечь P95 времени ответа
jq -s '[.[] | select(.metric == "http_req_duration")] | map(.value) | sort | .[. * 0.95 | floor]' results/catalog.json

# Извлечь rate ошибок
jq -s '[.[] | select(.metric == "http_req_failed")] | map(.value) | add / length' results/catalog.json
```

### HTML отчёт

Автоматически генерируется `index.html` с:
- Общей сводкой (pass/fail по threshold'ам)
- Карточками по каждой метрике
- Цветовой индикацией (зелёный = pass, красный = fail)

## Thresholds (auto-fail)

Все скрипты настроены с `thresholds` — k6 вернёт exit code 1 при нарушении SLA:

```javascript
thresholds: {
    http_req_duration: ['p(95)<200', 'p(99)<500'],  // P95 < 200ms, P99 < 500ms
    http_req_failed: ['rate<0.01'],                  // < 1% ошибок
}
```

## CI/CD интеграция

### GitHub Actions

```yaml
- name: Load testing
  run: |
    docker run -d --name lkfl-server lkfl/lkfl-server:latest
    sleep 10
    docker run --network host -v $(pwd)/loadtest:/scripts grafana/k6 run /scripts/run.sh
```

### GitLab CI

```yaml
load-test:
  image: grafana/k6:latest
  script:
    - k6 run --out json=results.json catalog.js
  artifacts:
    paths:
      - results.json
    reports:
      junit: results.xml
```

## Интерпретация результатов

### ✅ Прошёл все тесты

```
     ✓{r} http_req_duration..............: avg=45.2ms   min=12ms   med=38ms   max=234ms   p(90)=89ms   p(95)=112ms  p(99)=187ms
     ✓{r} http_req_failed................: 0.00% ✓{r} 0 out of 15234
```

Все threshold'ы зелёные, SLA выполнен.

### ❌ Не прошёл

```
     ✗{r} http_req_duration..............: avg=234ms    min=45ms   med=189ms  max=2340ms  p(90)=412ms  p(95)=567ms  p(99)=890ms
     ✓{r} http_req_failed................: 0.00% ✓{r} 0 out of 15234
```

P95 = 567ms > 200ms — threshold нарушен. k6 вернёт exit code 1.

### Что делать при провале

1. **Проверьте ресурсы сервера** — CPU, RAM, disk I/O
2. **Проверьте базу данных** — connection pool, slow queries
3. **Проверьте Redis** — hit rate, latency
4. **Увеличьте connection pool** в PostgreSQL
5. **Добавьте кэширование** для горячих запросов
6. **Профилируйте** — `go tool pprof` для Go backend

## Кастомизация

### Изменение URL цели

```bash
BASE_URL=http://production.lkfl.local ./loadtest/run.sh catalog
```

### Изменение нагрузки в скрипте

Измените `options.stages` в нужном JS файле:

```javascript
stages: [
    { duration: '1m', target: 1000 },  // 1000 VU вместо 500
    { duration: '2m', target: 1000 },
    { duration: '1m', target: 0 },
]
```

### Docker запуск

```bash
docker run -v $(pwd)/loadtest:/scripts -e BASE_URL=http://host.docker.internal:8080 grafana/k6 run /scripts/catalog.js
```

## Связь с задачами

| Задача | Описание |
|--------|----------|
| T2205 | Эта задача — k6 load testing |
| T2203 | Unit тесты (dependency) |
| T2204 | Интеграционные тесты |
| T2206 | CI/CD pipeline для load тестов |
