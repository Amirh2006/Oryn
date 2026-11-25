# ==============================================================================
# Build Stage
# ==============================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build tools and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compile static binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /app/bin/server ./cmd/server/main.go

# ==============================================================================
# Production Image
# ==============================================================================
FROM alpine:3.20

# Create non-root unprivileged user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy SSL certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/bin/server ./server

# Enforce non-root ownership
RUN chown -R appuser:appgroup /home/appuser
USER appuser

EXPOSE 8080

ENTRYPOINT ["./server"]
