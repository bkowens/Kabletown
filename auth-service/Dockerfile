FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations
CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /build/bin/server ./cmd/server

# Runtime image
FROM alpine:3.19

# Install runtime dependencies (ca-certificates for TLS, ca-certificates for HTTPS requests)
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/server /app/server

# Expose port
EXPOSE 8001

# Set user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8001/health || exit 1

CMD ["/app/server"]
