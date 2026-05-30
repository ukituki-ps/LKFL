# LKFL — Production Multi-stage Dockerfile
# Таргеты: server (lkfl-server), proxy (lkfl-integration-proxy)
#
# Build:
#   docker build --target server -t lkfl-server:latest .
#   docker build --target proxy  -t lkfl-integration-proxy:latest .
#
# Production images use distroless/base-debian12 (minimal, non-root).
# No HEALTHCHECK in image — distroless has no shell/wget.
# Healthcheck managed externally via docker-compose nginx upstream.

# ============================================================================
# Stage 0: Base — Go инструментал
# ============================================================================
FROM golang:1.24-alpine AS base

# Allow Go to download toolchain (needed for go.mod >= 1.25)
ENV GOTOOLCHAIN=auto

# Системные зависимости для сборки
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Кэширование слоёв зависимостей (go mod download до COPY исходников)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# ============================================================================
# Stage 1: Server — lkfl-server binary
# ============================================================================
FROM base AS server-build

COPY backend/ ./

# Сборка: CGO_ENABLED=0 (static binary), -trimpath (reproducible), -ldflags (strip)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /server \
    ./cmd/server/

# ============================================================================
# Stage 2: Proxy — lkfl-integration-proxy binary
# ============================================================================
FROM base AS proxy-build

COPY backend/ ./

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /proxy \
    ./cmd/integration-proxy/

# ============================================================================
# Stage 3: Frontend — pre-built dist/ from build context
# Frontend is built locally (npm run build) and dist/ is included.
# For CI with private npm registry: build frontend separately, mount dist/.
# ============================================================================
FROM scratch AS frontend-build

COPY frontend/dist/ /dist/

# ============================================================================
# Stage 4: Server runtime — lkfl-server (production)
#
# distroless/base-debian12: ~50MB, no shell, no wget, no package manager.
# Non-root user uid 1001. Healthcheck managed externally (docker-compose / orchestrator).
# ============================================================================
FROM gcr.io/distroless/base-debian12 AS server

WORKDIR /app

# Копируем бинарь, фронтенд dist, миграции
COPY --chown=1001:1001 --from=server-build /server /app/server
COPY --chown=1001:1001 --from=frontend-build /dist /app/dist
COPY --chown=1001:1001 migrations/ /app/migrations/

USER 1001:1001

EXPOSE 8080

# No HEALTHCHECK here — distroless has no shell/wget.
# Healthcheck is configured in docker-compose via nginx upstream check.

ENTRYPOINT ["/app/server"]

# ============================================================================
# Stage 5: Proxy runtime — lkfl-integration-proxy (production)
# ============================================================================
FROM gcr.io/distroless/base-debian12 AS proxy

WORKDIR /app

COPY --chown=1001:1001 --from=proxy-build /proxy /app/proxy

USER 1001:1001

EXPOSE 8090 8091

# No HEALTHCHECK — distroless has no shell/wget. Proxy is a stub in F1.

ENTRYPOINT ["/app/proxy"]
