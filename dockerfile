# =============================
# FRONTEND BUILD STAGE
# =============================
FROM node:20-alpine AS web-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

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

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /file-server /file-server
COPY --from=web-builder /app/frontend/dist /app/dist

# Data directory writable by the non-root runtime user.
RUN mkdir -p /data && chown -R nobody:nobody /data

WORKDIR /data
ENV APP_ENV=production \
    PORT=22010 \
    ROOT_DIR=/data \
    STATIC_DIR=/app/dist \
    ENABLE_TLS=false \
    ENABLE_AUTH=true

EXPOSE 22010

USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://127.0.0.1:22010/health || exit 1

ENTRYPOINT ["/file-server"]
