# ADR-037: Keycloak behind reverse proxy — один nginx, чистый verifier.go

| Поле     | Значение |
|----------|----------|
| Status   | Accepted |
| Date     | 2026-05-29 |
| Задача   | T2215 |
| Веха     | M22 (F1 Hardening) |
| Авторы   | architect-lkfl |

## Context

28-29 мая 2026 — серия проблем на staging с Keycloak login flow потребовала 12+ коммитов и 9 хак-решений:

1. `callback 500` → migration для `metadata`/`settings` колонок
2. `HTTP redirect` → `X-Forwarded-Proto` fix в nginx
3. `HTTP form action` → hostname:v2 вместо `KC_PROXY_ADDRESS_FORWARDING`
4. `HTTP discovery` → `KC_HOSTNAME_BACKCHANNEL_URL`
5. `Keycloak не читал X-Forwarded-Proto` → `KC_PROXY_HEADERS: xforwarded`
6. `HTTPS issuer потребовал SSL` → `SSL_CERT_FILE`, self-signed cert
7. `Revert` → потеря healthcheck + command
8. `Итог` → 3 хак-решения в Go коде, 9 файлов/настроек-костылей

### Корневая причина

**Двойной nginx:** `external nginx (TLS) → internal nginx (HTTP + HTTPS self-signed)`.

Internal nginx создал необходимость HTTPS внутри Docker-сети, что повлекло:
- Self-signed cert с SAN → Go не доверяет → `SSL_CERT_FILE` + `TLS_INSECURE`
- Подмена `http.DefaultTransport` глобально (гоночная ситуация)
- `extra_hosts` маппинг public URL на host-gateway
- `KC_PROXY_HEADERS` + `KC_HOSTNAME_STRICT_HTTPS` в Keycloak (конфликт по дизайну)
- 95 строк в verifier.go вместо 18

### Альтернативы

| Вариант | Плюсы | Минусы |
|---------|-------|--------|
| **Текущий (двойной nginx + HTTPS внутри)** | internal nginx route'рует всё | 12 коммитов костылей, гоночная ситуация в Go, не воспроизводит прод |
| **Один external nginx, три порта** | Простота, 0 костылей, чистый Go | Требует external nginx config на сервере |
| **Keycloak на HTTPS внутри** | issuer HTTPS | self-signed cert, Go TLS verification, не воспроизводит прод |
| **Traefik вместо nginx** | auto-discovery Docker labels | лишняя зависимость, обучение, не решает корневую проблему |

## Decision

**Один external nginx на сервере, три порта напрямую.**

```
                    Браузер
                       │ HTTPS (Let's Encrypt / self-signed)
                       ▼
            ┌──────────────────────────┐
            │  External Nginx          │  ← единственный reverse proxy
            │  (serverPr01)            │     TLS termination здесь
            │  443: ssl termination    │
            └───┬──────┬───────┬──────┘
                │      │       │  (всё HTTP внутри Docker-сети)
           18080│   19081│   8081│
                ▼      ▼       ▼
          lkfl-server keycloak lkfl-frontend
```

### Что убрано

| Файл/настройка | Было | Теперь |
|---------------|------|--------|
| internal nginx (container) | маршрутизация внутри Docker | ❌ удалён |
| `infra/nginx/ssl/` | self-signed cert | ❌ удалён |
| `verifier.go:newHTTPClient()` | подмена `http.DefaultTransport` | ❌ удалён |
| `docker-compose.staging.yml` — nginx | service | ❌ удалён |
| `docker-compose.staging.yml` — SSL_CERT_FILE | env | ❌ удалён |
| `docker-compose.staging.yml` — TLS_INSECURE | env | ❌ удалён |
| `docker-compose.staging.yml` — extra_hosts public | маппинг | ❌ удалён |
| `docker-compose.staging.yml` — volume server.crt | mount | ❌ удалён |
| `default.conf` — HTTPS server block | 443 ssl | ❌ удалён |
| `keycloak: KC_PROXY_HEADERS` | xforwarded | ❌ удалён |
| `keycloak: KC_HOSTNAME_STRICT_HTTPS` | "true" | ❌ удалён |

**11 файлов/настроек убрано.**

### verifier.go — до

```go
func newHTTPClient() *http.Client {
    pool, err := x509.SystemCertPool()
    if err != nil {
        pool = x509.NewCertPool()
    }
    certFile := os.Getenv("SSL_CERT_FILE")
    if certFile != "" {
        data, err := os.ReadFile(certFile)
        if err == nil && pool.AppendCertsFromPEM(data) {
            slog.Info("loaded custom CA cert", "file", certFile)
        }
    }
    if os.Getenv("TLS_INSECURE") == "true" {
        slog.Warn("TLS verification disabled (staging)")
        return &http.Client{
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
            },
        }
    }
    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{RootCAs: pool},
        },
    }
}

func NewVerifier(ctx context.Context, issuerURL, clientID string) (*oidc.IDTokenVerifier, error) {
    var provider *oidc.Provider
    var err error
    oldTransport := http.DefaultTransport
    http.DefaultTransport = newHTTPClient().Transport  // ← гонка
    defer func() { http.DefaultTransport = oldTransport }()
    for i := 0; i < 30; i++ {
        provider, err = oidc.NewProvider(ctx, issuerURL)
        if err == nil { break }
        time.Sleep(2 * time.Second)
    }
    if err != nil {
        return nil, fmt.Errorf("oidc provider (after 30 retries): %w", err)
    }
    return provider.Verifier(&oidc.Config{ClientID: clientID}), nil
}
```

### verifier.go — после

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

### Keycloak hostname:v2 — правильные настройки

| Настройка | Значение | Зачем |
|-----------|----------|-------|
| `KC_HOSTNAME` | `dev.april.ukituki.tech` | Browser redirect URI hostname |
| `KC_HOSTNAME_STRICT` | `"false"` | Не валидировать Host header |
| `KC_HOSTNAME_STRICT_HTTP` | `"false"` | Разрешить HTTP внутри Docker-сети |

| Настройка | ❌ НЕ использовать | Почему |
|-----------|-------------------|--------|
| `KC_PROXY_HEADERS` | xforwarded | Конфликт с hostname:v2, документация Keycloak не рекомендует |
| `KC_HOSTNAME_STRICT_HTTPS` | `"true"` | Превращает issuer в HTTPS → требует TLS verification внутри |
| `KC_PROXY_ADDRESS_FORWARDING` | — | Устарел, конфликтует с hostname:v2 |

### Keycloak — issuer на staging

```
issuer: http://keycloak:8080/realms/lkfl-sdek
iss claim в токене: http://keycloak:8080/realms/lkfl-sdek
```

**Это корректно.** go-oidc проверяет совпадение `iss` claim с `issuerURL` (без проверки scheme). Browser redirect URI берётся из `KC_HOSTNAME`, не из issuer.

### External nginx config (пример)

```nginx
upstream lkfl_backend  { server 127.0.0.1:18080; }
upstream lkfl_keycloak { server 127.0.0.1:19081; }
upstream lkfl_frontend { server 127.0.0.1:8081;  }

server {
    listen 443 ssl;
    server_name dev.april.ukituki.tech;

    ssl_certificate     ...;
    ssl_certificate_key ...;

    # Backend API
    location /api   { proxy_pass http://lkfl_backend;  }
    location /admin { proxy_pass http://lkfl_backend;  }

    # Keycloak OAuth
    location /realms { proxy_pass http://lkfl_keycloak; }
    location /protocol { proxy_pass http://lkfl_keycloak; }
    location /login-actions { proxy_pass http://lkfl_keycloak; }
    location /account { proxy_pass http://lkfl_keycloak; }
    location /resources { proxy_pass http://lkfl_keycloak; }
    location /services { proxy_pass http://lkfl_keycloak; }

    # Frontend SPA
    location / { proxy_pass http://lkfl_frontend; }
}
```

## Consequences

### Позитивные

- **0 костылей в Go коде** — verifier.go чистый, без `crypto/tls`, без подмены глобальных переменных
- **0 self-signed cert внутри** — TLS termination только на границе сети
- **1 nginx вместо 2** — меньше контейнеров, меньше конфигов
- **11 файлов/настроек убрано** — меньше surface area для ошибок
- **Повторяемость** — больше не нужно 12 коммитов для одной проблемы
- **Совпадение с продом** — на проде тоже external LB → 3 порта

### Негативные

- **Требует external nginx на сервере** — нужно поддерживать конфиг на serverPr01 (уже есть, просто обновить mapping)
- **Issuer claim в токене HTTP** — только на staging, на проде будет HTTPS (не влияет на функциональность)

## Правила (не нарушать)

### Правило 1: TLS termination — один слой

TLS termination происходит **только на границе сети** (external nginx / LB).
Внутри Docker-сети — **всегда HTTP**. Никаких self-signed cert'ов внутри.

### Правило 2: Keycloak issuer = внутренний URL

`KEYCLOAK_ISSUER` — URL для Go server discovery → всегда **внутренний HTTP URL**.
Browser redirect URI задаётся `KC_HOSTNAME`, не issuer.

### Правило 3: hostname:v2 без KC_PROXY_HEADERS

`KC_PROXY_HEADERS` конфликтует с hostname:v2. Использовать только `KC_HOSTNAME` + `KC_HOSTNAME_STRICT`.

### Правило 4: verifier.go — только go-oidc

`NewVerifier()` — чистая обёртка над `oidc.NewProvider()` с retry.
Никаких `crypto/tls`, `http.Transport`, `x509`, подмены глобальных переменных.
