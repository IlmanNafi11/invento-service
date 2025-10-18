FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags="-s -w" -o bin/app cmd/app/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/bin/app .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/queries ./queries
COPY --from=builder /app/.env.example .env

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:3000/health || exit 1

CMD ["./app"]
