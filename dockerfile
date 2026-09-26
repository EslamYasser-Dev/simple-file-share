# =============================
# FRONTEND BUILD STAGE
# =============================
FROM node:20-alpine AS web-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/yarn.lock frontend/.yarnrc.yml ./
RUN corepack enable && yarn install --immutable

COPY frontend/ ./
RUN yarn build

# =============================
# BACKEND BUILD STAGE
# =============================
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./backend
WORKDIR /app/backend

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /file-server ./cmd/server

# =============================
# RUNTIME STAGE
# =============================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates wget su-exec

COPY --from=builder /file-server /file-server
COPY --from=web-builder /app/frontend/dist /app/dist
COPY docker-entrypoint.sh /docker-entrypoint.sh

# Data directory writable by the non-root runtime user; the entrypoint also
# creates and owns ROOT_DIR at boot, then drops from root to nobody.
RUN mkdir -p /data && chown -R nobody:nobody /data && \
    chmod +x /docker-entrypoint.sh

# The server refuses ROOT_DIR == working directory, so cwd must not be /data.
WORKDIR /app
ENV APP_ENV=production \
    PORT=22010 \
    GRPC_PORT=50051 \
    ROOT_DIR=/data \
    STATIC_DIR=/app/dist \
    ENABLE_TLS=false \
    ENABLE_AUTH=true \
    ENABLE_GRPC=true \
    ENABLE_GRPC_TLS=true \
    CORS_ORIGINS=https://simple-file-share-eight.vercel.app,https://shares-iota.vercel.app

EXPOSE 22010 50051

# Starts as root only inside the entrypoint (mkdir/chown ROOT_DIR);
# su-exec drops the actual server process to nobody.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://127.0.0.1:22010/health || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
