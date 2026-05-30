# T2211 — Security Audit

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Безопасностный аудит F1.
OWASP Top 10 проверка, rate limiting, CORS, dependency audit, penetration testing.

## Что сделать

### OWASP Top 10 проверка

1. **A01: Broken Access Control**
   - Vertical privilege escalation: employee → admin routes
   - Horizontal privilege escalation: tenant A → tenant B data
   - IDOR: direct object reference на user_id, engagement_id

2. **A02: Cryptographic Failures**
   - JWT signature verification
   - Password hashing (если применяется)
   - TLS everywhere (no HTTP)

3. **A03: Injection**
   - SQL injection: параметризованные запросы (pgx, no string formatting)
   - No raw SQL strings

4. **A05: Security Misconfiguration**
   - Default credentials removed
   - Debug mode disabled in production
   - Verbose error messages disabled
   - CORS policy: allow only staging domain

5. **A07: Identification and Authentication Failures**
   - Session management: token expiration, refresh token rotation
   - Brute force protection: rate limiting на login

6. **A09: Security Logging and Monitoring Failures**
   - Auth events logged (login, logout, failure)
   - Suspicious activity detection

### Rate limiting

- Auth endpoints: 10 requests/minute per IP
- Catalog endpoints: 100 requests/minute per IP
- Admin endpoints: 60 requests/minute per IP
- Implementation: Redis-based sliding window

### CORS policy

- Allowed origins: staging domain only
- Allowed methods: GET, POST, PATCH, DELETE
- Allowed headers: Authorization, Content-Type, X-Tenant-ID
- Credentials: true (for cookies if used)
- Preflight caching: 1 hour

### Dependency audit

- **Go:** `govulncheck ./...` — 0 vulnerabilities
- **Frontend:** `npm audit` — 0 high/critical vulnerabilities
- **Docker:** `trivy image lkfl-server:latest` — 0 critical

### Penetration testing

- SQL injection test: параметризованные запросы, pgx
- XSS test: frontend escaping, CSP headers
- CSRF test: SameSite cookies, CSRF token if applicable
- SSRF test: no external URL fetching from server

## Критерии приёмки

- [ ] OWASP Top 10 check пройден
- [ ] Rate limiting на auth endpoints (10 req/min)
- [ ] Rate limiting на catalog endpoints (100 req/min)
- [ ] CORS policy verification
- [ ] govulncheck — 0 vulnerabilities
- [ ] npm audit — 0 high/critical
- [ ] trivy image — 0 critical
- [ ] SQL injection test пройден
- [ ] XSS test пройден
