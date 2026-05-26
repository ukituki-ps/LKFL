# T1005 — Shared auth pkg для payment-gateway

## Веха

M10-рефакторинг-по-результатам-аудита

## Контекст

`payment-gateway/internal/auth/` документируется как "same as platform" (модули.md стр. 304):
```
auth/ | JWT validation (same as platform) | VerifyToken(), RBACGuard() | go-oidc, Redis
```

Это означает дублирование кода между:
- `platform/internal/auth/` — OIDC verifier, JWT middleware, RBAC guard
- `payment-gateway/internal/auth/` — копия того же функционала

**Проблема:**
Monorepo (ADR-011). Два сервиса имеют одинаковый go-oidc client + middleware logic. Изменение в одном (например, добавление tenant claim validation) требует copy-paste в другой.

**Решение — shared/pkg/auth:**
Вынести OIDC verification + middleware в shared Go package:
```
shared/
  pkg/
    auth/
      verifier.go     # OIDC verifier, JWT validation
      middleware.go   # JWTMiddleware, TenantResolver
      rbac.go         # RBACGuard role check
```

```go
// payment-gateway/internal/... imports shared/pkg/auth directly
// platform/internal/auth/ becomes alias or also uses shared/pkg/auth
```

Platform `auth/` может остаться как thin wrapper (tenant-specific config), но core logic = shared.

### Файлы-мишени

| Действие | Файл |
|---|--|
| shared/pkg/auth | `архитектура/модули.md` — Payment Gateway auth/ → shared/pkg/auth |
| platform auth/ | `архитектура/пакеты-platform.md` — auth/ использует shared + tenant wrapper |
| ADR-011 | `архитектура/adr/011-monorepo.md` — shared/ section (cel-context + auth = 2 shared pkgs) |

### Критерии приёмки

- [ ] `архитектура/модули.md` — payment-gateway auth/ → ссылка на shared/pkg/auth
- [ ] shared/pkg/auth задокументирован: verifier.go, middleware.go, rbac.go
- [ ] Platform auth/ использует shared + thin tenant wrapper в `пакеты-platform.md`
- [ ] "same as platform" text → "uses shared/pkg/auth" в `модули.md`
- [ ] ADR-011 monorepo — shared/ expanded (cel-context + auth). **Зависит от T1003:** ADR-011 обновлен T1003 (cel-context), затем T1005 добавляет auth
