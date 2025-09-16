# =============================
# BUILD STAGE
# =============================
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install Git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source
COPY . .

# Build binary with flags for smaller, production-ready binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o /file-server ./cmd/server

# =============================
# FINAL STAGE (SCRATCH)
# =============================
FROM scratch

# Copy binary
COPY --from=builder /file-server /file-server

# Create non-root user (UID/GID 65534 = nobody on Alpine)
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid 65534 \
    nobody

VOLUME ["/data"]

# Expose port
EXPOSE 22010

# Run as non-root
USER nobody

# Healthcheck (we'll implement /health endpoint soon!)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:22010/health || exit 1

# Entrypoint
ENTRYPOINT ["/file-server"]