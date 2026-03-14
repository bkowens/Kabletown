# Metadata Service Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies for MySQL driver
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o metadata-service ./cmd/server

# Runtime stage
FROM alpine:3.18
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/metadata-service .

EXPOSE 8008

ENTRYPOINT ["./metadata-service"]
