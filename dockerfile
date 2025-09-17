# =============================
# BUILD STAGE
# =============================
FROM golang:latest-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o /file-server ./cmd/server

# =============================
# FINAL STAGE (SCRATCH)
# =============================
FROM scratch

COPY --from=builder /file-server /file-server

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid 65534 \
    nobody

VOLUME ["/data"]

EXPOSE 22010

USER nobody

# Healthcheck (we'll implement /health endpoint soon!)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:22010/health || exit 1

# Entrypoint
ENTRYPOINT ["/file-server"]