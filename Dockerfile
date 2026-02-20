# =============================================================================
# Stage 1: Build
# =============================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Cache Go module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (respects .dockerignore)
COPY . .

# Build statically-linked binary with stripped debug info
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app ./cmd/app

# =============================================================================
# Stage 2: Runtime
# =============================================================================
FROM alpine:3.21

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /build/app .

# Create uploads directory for local file storage (TUS resumable uploads)
RUN mkdir -p uploads && \
    chown -R appuser:appuser /app

USER appuser

# Port configurable via environment variable (override in Dokploy/runtime)
ENV APP_PORT=3000
EXPOSE ${APP_PORT}

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD sh -c "wget --no-verbose --tries=1 --spider http://localhost:${APP_PORT}/api/v1/health || exit 1"

CMD ["./app"]
