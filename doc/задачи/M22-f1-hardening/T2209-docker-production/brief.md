# T2209 — Docker Production

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Production-ready Docker образы для lkfl-server и lkfl-integration-proxy.
Multi-stage build, безопасность, оптимизация размера.

## Что сделать

### lkfl-server Dockerfile

**Multi-stage build:**
1. **build** — golang:1.23-alpine, CGO_ENABLED=0, go build
2. **frontend** — node:20-alpine, npm ci, npm run build
3. **production** — scratch/distroless, binary + frontend dist + config

**Безопасность:**
- Non-root user (uid 1001)
- Read-only filesystem (`--read-only`)
- Temporary filesystem для `/tmp`
- No shell in production image
- Minimal base (scratch или distroless-static)

**Healthcheck:**
- `HEALTHCHECK CMD wget -q --spider http://localhost:8080/healthz || exit 1`
- Interval: 30s, Timeout: 10s, Retries: 3

### lkfl-integration-proxy Dockerfile

Аналогично multi-stage, non-root, healthcheck.

### Image signing (cosign)

- Подпись образов: `cosign sign lkfl-server:tag`
- Верификация при pull: `cosign verify`
- Key management: Kubernetes secret или cosign key pair

### .dockerignore

- Исключить: `node_modules/`, `.git/`, `doc/`, `*.md`, `test/`, `loadtest/`

### Docker Compose production

- `docker-compose.prod.yml` — production configuration
- Resource limits: memory, CPU
- Restart policy: `unless-stopped`
- Health dependencies: `depends_on` с condition

## Критерии приёмки

- [ ] lkfl-server Dockerfile (multi-stage, distroless)
- [ ] lkfl-integration-proxy Dockerfile (multi-stage)
- [ ] Non-root user (uid 1001)
- [ ] Read-only filesystem
- [ ] Healthcheck endpoint
- [ ] Image signing (cosign)
- [ ] .dockerignore настроен
- [ ] docker-compose.prod.yml
- [ ] Размер образа < 50MB (lkfl-server)
