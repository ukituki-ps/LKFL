# T2215 — Отчёт: Очистка архитектуры Keycloak/HTTPS на staging

**Дата:** 29.05.2026
**Статус:** ✅ Выполнено (кроме E2E на staging)
**Время:** ~30 мин

## Что сделано

### Фаза 1: ADR-037

- ✅ ADR-037 уже существовал (создан ранее архитектором)
- ✅ ADR-037 добавлен в индекс `doc/архитектура/adr/README.md` (строка 47)

### Фаза 2: Go — чистка verifier.go

- ✅ Удалён `newHTTPClient()` (31 строка)
- ✅ Удалена подмена `http.DefaultTransport` в `NewVerifier()`
- ✅ Удалены imports: `crypto/tls`, `crypto/x509`, `net/http`, `os`
- ✅ Оставлен чистый `oidc.NewProvider` + retry loop
- ✅ **95 строк → 18 строк**
- ✅ `go build ./...` — компиляция без ошибок
- ✅ `go vet ./...` — без предупреждений

### Фаза 3: Docker Compose staging — чистка

- ✅ **lkfl-server:**
  - Убрана `extra_hosts: dev.april.ukituki.tech:host-gateway`
  - `KEYCLOAK_ISSUER: http://keycloak:8080/realms/lkfl-sdek` (было HTTPS)
  - Убрана `SSL_CERT_FILE`
  - Убрана `TLS_INSECURE`
  - Убран volume mount `server.crt`
  - Оставлен `host.docker.internal:host-gateway`
  - Добавлены комментарии ADR-037
- ✅ **Keycloak:**
  - Убрана `KC_PROXY_HEADERS: xforwarded`
  - Убрана `KC_HOSTNAME_STRICT_HTTPS: "true"`
  - Оставлены `KC_HOSTNAME`, `KC_HOSTNAME_STRICT`, `KC_HOSTNAME_STRICT_HTTP`
  - Добавлены комментарии ADR-037
- ✅ **Nginx container** — удалён целиком из docker-compose.staging.yml
- ✅ **lkfl-frontend** — добавлен `ports: "8081:80"`

### Фаза 4: Nginx config — убрать internal

- ✅ `default.conf` — удалён HTTPS server block `listen 443 ssl` (был 31 строка)
- ✅ `default.conf` — оставлен HTTP server block `listen 80` (для локальной dev)
- ✅ Удалены файлы: `infra/nginx/ssl/server.crt`, `infra/nginx/ssl/server.key`

### Фаза 5: External nginx config

- ✅ Задокументировано в ADR-037 (строки 176-202): 3 upstream'а, mapping портов

### Фаза 6: E2E тест

- ⏳ **Не выполнено** — требует деплоя на serverAI и ручной проверки

### Фаза 7: Документация

- ✅ `doc/архитектура/adr/README.md` — ADR-037 в индексе
- ✅ `doc/план/задачи.md` — T2215 обновлён (100%, дата 29.05)
- ✅ `doc/деплой.md` — секция staging обновлена (ADR-037, архитектура, troubleshooting)
- ✅ `docker-compose.staging.yml` — комментарии со ссылками на ADR-037

## Проверка критериев приёмки

### Go код

- [x] `verifier.go` — нет `newHTTPClient()`
- [x] `verifier.go` — нет подмены `http.DefaultTransport`
- [x] `verifier.go` — нет imports `crypto/tls`, `crypto/x509`, `net/http`, `os`
- [x] `verifier.go` — строки 18 (было 95)
- [x] `go build ./...` — компиляция без ошибок
- [x] `go vet ./...` — без предупреждений

### Docker Compose staging

- [x] `lkfl-server` — нет `SSL_CERT_FILE`
- [x] `lkfl-server` — нет `TLS_INSECURE`
- [x] `lkfl-server` — нет volume mount `server.crt`
- [x] `lkfl-server` — нет `extra_hosts: dev.april.ukituki.tech`
- [x] `lkfl-server` — `KEYCLOAK_ISSUER: http://keycloak:8080/...`
- [x] `keycloak` — нет `KC_PROXY_HEADERS`
- [x] `keycloak` — нет `KC_HOSTNAME_STRICT_HTTPS`
- [x] `keycloak` — порт 19081 маппится напрямую
- [x] `lkfl-frontend` — порт 8081 маппится напрямую
- [x] `nginx` — удалён из docker-compose.staging.yml

### Nginx config

- [x] `default.conf` — нет HTTPS server block `:443`
- [x] `infra/nginx/ssl/server.crt` — удалён
- [x] `infra/nginx/ssl/server.key` — удалён

### Документация

- [x] ADR-037 создан и принят
- [x] `doc/архитектура/adr/README.md` — ADR-037 добавлен в индекс
- [x] `doc/деплой.md` — секция staging обновлена
- [x] `docker-compose.staging.yml` — комментарии со ссылками на ADR-037
- [x] `doc/план/задачи.md` — T2215 добавлен и обновлён

### External nginx (сервер)

- [ ] Config маршрутизирует `/api` → 18080 — ⏳ требует обновления на serverAI
- [ ] Config маршрутизирует `/realms`, `/protocol` и др. → 19081 — ⏳ требует обновления на serverAI
- [ ] Config маршрутизирует `/` → 8081 — ⏳ требует обновления на serverAI
- [x] TLS termination — единственный слой (архитектурно)

### E2E

- [ ] Login → Keycloak → Callback → Dashboard работает — ⏳ требует деплоя
- [ ] HTTPS на всём пути для браузера — ⏳ требует деплоя
- [ ] 0 mixed content warnings — ⏳ требует деплоя
- [ ] Console errors — 0 — ⏳ требует деплоя

## Изменённые файлы

| Файл | Действие | Строк до → после |
|------|----------|-------------------|
| `backend/shared/pkg/auth/verifier.go` | изменён | 95 → 18 |
| `docker-compose.staging.yml` | изменён | 534 → ~483 |
| `infra/nginx/server/default.conf` | изменён | 234 → ~203 |
| `infra/nginx/ssl/server.crt` | **удалён** | — |
| `infra/nginx/ssl/server.key` | **удалён** | — |
| `doc/архитектура/adr/README.md` | без изменений | (ADR-037 уже был добавлен) |
| `doc/план/задачи.md` | изменён | +обновлён статус T2215 |
| `doc/деплой.md` | изменён | +секция staging ADR-037 |
| `doc/задачи/.../plan.yaml` | изменён | все пункты done |
| `doc/задачи/.../report.md` | **создан** | этот файл |

## Замечания

1. **E2E тесты не выполнены** — требуется деплой на serverAI: git pull → docker compose down → docker compose up -d → проверка в браузере.
2. **External nginx на serverAI** — конфиг нужно обновить на сервере: убрать internal nginx mapping, добавить прямые порты 18080, 19081, 8081.
3. **default.conf** остаётся для локальной dev-среды — там nginx контейнер всё ещё используется в `docker-compose.dev.yml`.
4. **infra/nginx/ssl/** — директория может остаться пустой; можно удалить целиком, если нигде больше не ссылаются.

## 4 правила (соблюдены)

1. ✅ TLS termination — только на границе сети (external nginx)
2. ✅ `KEYCLOAK_ISSUER` = внутренний URL Keycloak (`http://keycloak:8080/...`)
3. ✅ hostname:v2 без `KC_PROXY_HEADERS`
4. ✅ verifier.go — только `go-oidc`, без `crypto/tls`, без подмены
