# T2215 — Очистка архитектуры Keycloak/HTTPS на staging

## Веха

M22-f1-hardening

## Тип

code + infra + docs

## Контекст

### История проблемы

28-29 мая 2026 — серия проблем на staging (`dev.april.ukituki.tech`) с Keycloak login flow:

1. **callback 500** — колонки `metadata`/`settings` не были в миграциях → добавлены миграции
2. **HTTP redirect** — nginx передавал `$scheme` (HTTP) в `X-Forwarded-Proto` → заменили на жёсткое `https`
3. **HTTP form action** — `KC_PROXY_ADDRESS_FORWARDING` конфликтовал с hostname:v2 → заменили на hostname:v2
4. **HTTP discovery issuer** — добавили `KC_HOSTNAME_BACKCHANNEL_URL`
5. **Keycloak не читал X-Forwarded-Proto** → добавили `KC_PROXY_HEADERS: xforwarded`
6. **HTTPS issuer потребовал SSL verification** → добавили `SSL_CERT_FILE`, nginx:443 с self-signed cert
7. **Revert** — убрали HTTPS issuer, вернулись к HTTP — но потеряли healthcheck и `command:`
8. **Итог** — 12+ коммитов, 3 хак-решения в Go коде (SSL_CERT_FILE, TLS_INSECURE, подмена http.DefaultTransport)

### Корневая причина

**Двойной nginx** — `external nginx (TLS termination) → internal nginx (HTTP + HTTPS self-signed)`.
Это создало необходимость HTTPS внутри Docker-сети, что повлекло цепочку костылей.

```
                    Браузер
                       │ HTTPS (реальный, Let's Encrypt)
                       ▼
            ┌─────────────────────┐
            │  External Nginx     │  ← TLS termination здесь
            │  (serverPr01)       │
            └─────────┬───────────┘
                      │ HTTP (внутри сервера)
                      ▼
            ┌─────────────────────┐
            │  Internal Nginx     │  ← ЛИШНИЙ СЛОЙ
            │  (docker container) │     терминирует self-signed TLS на 443
            │  port 80 + 443      │     для discovery
            └───┬───────────┬─────┘
                │            │
                ▼            ▼
         lkfl-server    keycloak
         (Go)           (HTTP)
```

### Что сейчас (костыли)

| Файл/настройка | Костыль | Зачем создан |
|---------------|---------|-------------|
| `verifier.go:28-58` | `newHTTPClient()` — подменяет `http.DefaultTransport` глобально | self-signed cert trust |
| `docker-compose.staging.yml:181` | `SSL_CERT_FILE: /etc/nginx/ssl/server.crt` | дать cert Go контейнеру |
| `docker-compose.staging.yml:182` | `TLS_INSECURE: "true"` | fallback если cert не загрузился |
| `docker-compose.staging.yml:184` | volume mount `server.crt` в lkfl-server | дать cert контейнеру |
| `docker-compose.staging.yml:160` | `extra_hosts: dev.april.ukituki.tech:host-gateway` | маппинг public URL для discovery |
| `docker-compose.staging.yml:168` | `KEYCLOAK_ISSUER: https://...` | HTTPS issuer для discovery |
| `docker-compose.staging.yml:115` | `KC_PROXY_HEADERS: xforwarded` | заставить Keycloak видеть HTTPS из nginx headers |
| `docker-compose.staging.yml:114` | `KC_HOSTNAME_STRICT_HTTPS: "true"` | принудить HTTPS issuer |
| `infra/nginx/server/default.conf:27-57` | HTTPS server block `:443` с self-signed | discovery endpoint по HTTPS |
| `infra/nginx/ssl/` | self-signed cert + key | TLS для internal nginx |

**9 файлов/настроек-костылей.**

### Почему это плохо

1. **Подмена `http.DefaultTransport`** — гоночная ситуация: на 60 сек (30 × 2s retries) весь процесс использует кастомный транспорт. Другой goroutine делает HTTP-запрос → получит модифицированный клиент.
2. **Self-signed cert на staging** — не воспроизводит прод. На проде будет реальный cert → тесты ложнопроходят.
3. **95 строк в verifier.go** вместо 18 — `crypto/tls`, `crypto/x509`, `net/http`, `os` в imports.
4. **Double nginx** — лишний слой маршрутизации, лишний контейнер, лишний конфиг для поддержки.
5. **`KC_PROXY_HEADERS` + `STRICT_HTTPS`** — конфликт по дизайну с hostname:v2. Документация Keycloak явно не рекомендует их совместно.

## Решение

### Архитектура

**Один external nginx, три порта напрямую:**

```
                    Браузер
                       │ HTTPS (Let's Encrypt / self-signed)
                       ▼
            ┌──────────────────────────┐
            │  External Nginx          │  ← единственный reverse proxy
            │  (serverPr01)            │
            │  443: ssl termination    │
            └───┬──────┬───────┬──────┘
                │      │       │  (всё HTTP внутри)
           18080│   19081│   8081│
                ▼      ▼       ▼
          lkfl-server keycloak frontend
```

**Internal nginx — удаляется из staging.**

### Keycloak — разделить frontend URL и issuer

```yaml
keycloak:
  ports:
    - "19081:8080"  # напрямую через external nginx
  environment:
    KC_HOSTNAME: dev.april.ukituki.tech
    KC_HOSTNAME_STRICT: "false"
    KC_HOSTNAME_STRICT_HTTP: "false"
    # Без KC_PROXY_HEADERS, без KC_HOSTNAME_STRICT_HTTPS
    # issuer → http://keycloak:8080/realms/lkfl-sdek (внутренний)
    # form action, redirect_uri → https://dev.april.ukituki.tech (из KC_HOSTNAME)
```

### lkfl-server — прямой HTTP к Keycloak

```yaml
lkfl-server:
  # Без extra_hosts маппинга
  # Без SSL_CERT_FILE
  # Без TLS_INSECURE
  # Без volume mount server.crt
  environment:
    KEYCLOAK_ISSUER: http://keycloak:8080/realms/lkfl-sdek
```

### External nginx — маршрутизация по портам

```nginx
upstream lkfl_backend  { server 127.0.0.1:18080; }
upstream lkfl_keycloak { server 127.0.0.1:19081; }
upstream lkfl_frontend { server 127.0.0.1:8081;  }

server {
    listen 443 ssl;
    server_name dev.april.ukituki.tech;

    ssl_certificate     ...;
    ssl_certificate_key ...;

    location /api   { proxy_pass http://lkfl_backend;  }
    location /admin { proxy_pass http://lkfl_backend;  }
    location /realms { proxy_pass http://lkfl_keycloak; }
    location /protocol { proxy_pass http://lkfl_keycloak; }
    location /login-actions { proxy_pass http://lkfl_keycloak; }
    location /account { proxy_pass http://lkfl_keycloak; }
    location /resources { proxy_pass http://lkfl_keycloak; }
    location /services { proxy_pass http://lkfl_keycloak; }
    location / { proxy_pass http://lkfl_frontend; }
}
```

### verifier.go — убрать костыли

```go
func NewVerifier(ctx context.Context, issuerURL, clientID string) (*oidc.IDTokenVerifier, error) {
    var provider *oidc.Provider
    var err error

    for i := 0; i < 30; i++ {
        provider, err = oidc.NewProvider(ctx, issuerURL)
        if err == nil { break }
        slog.Warn("oidc provider not ready, retrying", "attempt", i+1, "error", err)
        time.Sleep(2 * time.Second)
    }
    if err != nil {
        return nil, fmt.Errorf("oidc provider (after 30 retries): %w", err)
    }
    return provider.Verifier(&oidc.Config{ClientID: clientID}), nil
}
```

**95 строк → 18 строк.**

### Issuer claim в токене — HTTP на staging

```
iss: "http://keycloak:8080/realms/lkfl-sdek"
```

Это **корректно для staging**. go-oidc проверяет совпадение `iss` claim с переданным `issuerURL`, не проверяет scheme. На продакшене оба будут HTTPS с реальными cert'ами.

## Зависимости

- **T2210** (Деплой на стенд) — staging docker-compose уже создан, нужно адаптировать
- **T2213** (CI/CD Deploy Worker) — deploy-worker уже работает, адаптировать конфиг

## Архитектура

ADR-037 (создаётся в рамках этой задачи).

---

## Что сделать

### Фаза 1: ADR

#### 1.1 ADR-037 — Keycloak behind reverse proxy

`doc/архитектура/adr/037-keycloak-reverse-proxy.md`:
- Status: Accepted
- Context: двойной nginx → 12+ коммитов костылей, 95 строк хак-кода в verifier.go
- Decision: один external nginx, три порта напрямую, internal nginx убран
- Consequences: чистый verifier.go, HTTP issuer внутри сети, HTTPS только на границе

### Фаза 2: Go — чистка verifier.go

#### 2.1 Убрать newHTTPClient()

`backend/shared/pkg/auth/verifier.go`:
- Удалить `newHTTPClient()` (строки 28-58)
- Удалить подмену `http.DefaultTransport` в `NewVerifier()`
- Удалить imports: `crypto/tls`, `crypto/x509`, `net/http`, `os`
- Оставить чистый `oidc.NewProvider` + retry loop

#### 2.2 Проверка компиляции

```bash
cd backend && go build ./...
```

### Фаза 3: Docker Compose staging — чистка

#### 3.1 lkfl-server

`docker-compose.staging.yml`:
```diff
   lkfl-server:
     extra_hosts:
-      - "dev.april.ukituki.tech:host-gateway"
       - "host.docker.internal:host-gateway"
     environment:
-      KEYCLOAK_ISSUER: https://dev.april.ukituki.tech/realms/lkfl-sdek
+      KEYCLOAK_ISSUER: http://keycloak:8080/realms/lkfl-sdek
-      SSL_CERT_FILE: /etc/nginx/ssl/server.crt
-      TLS_INSECURE: "true"
     volumes:
-      - ./infra/nginx/ssl/server.crt:/etc/nginx/ssl/server.crt:ro
```

#### 3.2 Keycloak

```diff
   keycloak:
     ports:
       - "19081:8080"
     environment:
-      KC_PROXY_HEADERS: xforwarded
-      KC_HOSTNAME_STRICT_HTTPS: "true"
       KC_HOSTNAME: dev.april.ukituki.tech
       KC_HOSTNAME_STRICT: "false"
       KC_HOSTNAME_STRICT_HTTP: "false"
```

#### 3.3 Nginx container — убрать

```diff
   nginx:
-    (удаляется из docker-compose.staging.yml)
```

#### 3.4 Frontend — открыть порт

```diff
   lkfl-frontend:
+    ports:
+      - "8081:80"
```

### Фаза 4: Nginx config — убрать internal

#### 4.1 Убрать HTTPS server block

`infra/nginx/server/default.conf`:
```diff
-  server {
-      listen 443 ssl;
-      ... (строки 27-57)
-  }
```

**Примечание:** `default.conf` остаётся для локальной dev-среды. На staging routing делает external nginx на сервере.

#### 4.2 Убрать self-signed cert

```diff
-  infra/nginx/ssl/server.crt
-  infra/nginx/ssl/server.key
```

#### 4.3 Обновить README/комментарии

Комментарии в `docker-compose.staging.yml` — задокументировать:
- Почему `KEYCLOAK_ISSUER: http://keycloak:8080/...` (ссылка на ADR-037)
- Почему нет `SSL_CERT_FILE` / `TLS_INSECURE`
- Как маршрутизация работает через external nginx

### Фаза 5: External nginx config (документация)

#### 5.1 Задокументировать external nginx config

`infra/nginx/external/README.md` (или в ADR-037):
- Конфиг external nginx для staging
- Конфиг external nginx для production (placeholder)
- Mapping портов: backend 18080, keycloak 19081, frontend 8081
- Как добавить новые upstream'ы

### Фаза 6: E2E тест

#### 6.1 Проверка login flow

```bash
# На staging
curl -k https://dev.april.ukituki.tech/healthz
# Login flow через browser (Playwright E2E)
```

### Фаза 7: Документация

#### 7.1 Обновить doc/архитектура/безопасность.md

Секция "TLS" — задокументировать:
- TLS termination — только на external nginx (граница сети)
- Внутри Docker-сети — HTTP
- Keycloak issuer — HTTP внутри, HTTPS на границе

#### 7.2 Обновить doc/деплой.md

Секция staging:
- Архитектура: external nginx → 3 порта
- Почему нет internal nginx
- Ссылка на ADR-037

#### 7.3 Обновить ADR index

`doc/архитектура/adr/README.md` — добавить ADR-037

## Критерии приёмки

### Go код

- [ ] `verifier.go` — нет `newHTTPClient()`
- [ ] `verifier.go` — нет подмены `http.DefaultTransport`
- [ ] `verifier.go` — нет imports `crypto/tls`, `crypto/x509`, `net/http`, `os`
- [ ] `verifier.go` — строки ~18 (было 95)
- [ ] `go build ./...` — компиляция без ошибок
- [ ] `go vet ./...` — без предупреждений

### Docker Compose staging

- [ ] `lkfl-server` — нет `SSL_CERT_FILE`
- [ ] `lkfl-server` — нет `TLS_INSECURE`
- [ ] `lkfl-server` — нет volume mount `server.crt`
- [ ] `lkfl-server` — нет `extra_hosts: dev.april.ukituki.tech`
- [ ] `lkfl-server` — `KEYCLOAK_ISSUER: http://keycloak:8080/...`
- [ ] `keycloak` — нет `KC_PROXY_HEADERS`
- [ ] `keycloak` — нет `KC_HOSTNAME_STRICT_HTTPS`
- [ ] `keycloak` — порт 19081 маппится напрямую
- [ ] `lkfl-frontend` — порт 8081 маппится напрямую
- [ ] `nginx` — удалён из docker-compose.staging.yml (или оставлен только для dev)

### Nginx config

- [ ] `default.conf` — нет HTTPS server block `:443`
- [ ] `default.conf` — нет self-signed cert volumes
- [ ] `infra/nginx/ssl/` — удалён (не нужен на staging)

### External nginx (сервер)

- [ ] Config маршрутизирует `/api` → 18080
- [ ] Config маршрутизирует `/realms`, `/protocol`, `/login-actions`, `/account`, `/resources`, `/services` → 19081
- [ ] Config маршрутизирует `/` → 8081
- [ ] TLS termination — единственный слой

### E2E

- [ ] Login → Keycloak → Callback → Dashboard работает
- [ ] HTTPS на всём пути для браузера
- [ ] 0 mixed content warnings
- [ ] Console errors — 0

### Документация

- [ ] ADR-037 создан и принят
- [ ] `doc/архитектура/adr/README.md` — ADR-037 добавлен в индекс
- [ ] `doc/деплой.md` — секция staging обновлена
- [ ] `docker-compose.staging.yml` — комментарии со ссылками на ADR-037
- [ ] `doc/план/задачи.md` — T2215 добавлен

### Git

- [ ] Один коммит: `refactor(staging): убрать двойной nginx, почистить Keycloak HTTPS архитектуру`
- [ ] Все изменения в одном коммите (не 12 отдельных)

---

## Почему больше не будет 2 дней по кругу

### Правило 1: TLS termination — один слой

TLS termination происходит **только на границе сети** (external nginx / LB).
Внутри Docker-сети — **всегда HTTP**. Никаких self-signed cert'ов внутри.

### Правило 2: Keycloak issuer = внутренний URL

`KEYCLOAK_ISSUER` — это URL, по которому Go сервер делает discovery.
Он всегда равен **внутреннему URL Keycloak** (`http://keycloak:8080/realms/...`).
Browser redirect URI задаётся `KC_HOSTNAME`, не issuer.

### Правило 3: hostname:v2 без KC_PROXY_HEADERS

| Настройка | Значение | Зачем |
|-----------|----------|-------|
| `KC_HOSTNAME` | `dev.april.ukituki.tech` | Browser redirect URI |
| `KC_HOSTNAME_STRICT` | `"false"` | Не валидировать Host header |
| `KC_HOSTNAME_STRICT_HTTP` | `"false"` | Разрешить HTTP внутри |
| `KC_HOSTNAME_STRICT_HTTPS` | **не использовать** | ломает issuer |
| `KC_PROXY_HEADERS` | **не использовать** | конфликт с hostname:v2 |

### Правило 4: verifier.go — только go-oidc

`NewVerifier()` — чистая обёртка над `oidc.NewProvider()` с retry.
Без `crypto/tls`, без `http.Transport`, без `x509`, без подмены глобальных переменных.
