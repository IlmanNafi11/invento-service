FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /build/app .
COPY --from=builder /build/migrations ./migrations
COPY --from=builder /build/queries ./queries
COPY --from=builder /build/api ./api

RUN mkdir -p uploads keys && \
    chown -R appuser:appuser /app

USER appuser

ENV GOMEMLIMIT=350MiB
ENV GOGC=100

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

CMD ["./app"]
