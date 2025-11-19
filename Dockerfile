# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies (gcc, musl-dev for CGO)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev

# Copy go module files
COPY go/go.mod go/go.sum ./
RUN go mod download

# Copy source code
COPY go/ .

# Build CLI and API binaries
RUN CGO_ENABLED=1 go build -o bin/tm ./cmd/cli
RUN CGO_ENABLED=1 go build -o bin/tm-web ./cmd/web

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite-libs \
    tzdata

# Copy binaries from builder
COPY --from=builder /build/bin/tm /usr/local/bin/tm
COPY --from=builder /build/bin/tm-web /usr/local/bin/tm-web

# Create data directories
RUN mkdir -p /data /config /logs

# Set environment variables
ENV TELOS_FILE=/config/telos.md

# Default command (CLI)
ENTRYPOINT ["tm"]
CMD ["--help"]
