FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install git and certificates for module downloads
RUN apk add --no-cache git ca-certificates

# Leverage Docker layer caching by downloading dependencies first
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Use BuildKit cache mounts to speed up compilation and set CGO_ENABLED=0 for static builds
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

FROM alpine:latest

WORKDIR /root/

# Install CA certs for outbound HTTPS calls (ES, etc.)
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]
