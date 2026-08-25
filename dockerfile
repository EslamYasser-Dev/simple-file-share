# =============================
# BUILD STAGE
# =============================
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /file-server ./cmd/server

# =============================
# RUNTIME STAGE
# =============================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /file-server /file-server

WORKDIR /data
VOLUME ["/data"]

ENV APP_ENV=production \
    PORT=22010 \
    ROOT_DIR=/data \
    ENABLE_TLS=false \
    ENABLE_AUTH=true

EXPOSE 22010

USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://127.0.0.1:22010/health || exit 1

ENTRYPOINT ["/file-server"]
