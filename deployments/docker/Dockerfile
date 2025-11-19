# Multi-stage build for Telos Matrix (Go implementation)
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o tm ./cmd/cli

# Runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata sqlite

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/tm .

# Create data directory
RUN mkdir -p /app/data

# Expose port for API server (if implemented)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./tm health || exit 1

ENTRYPOINT ["./tm"]
