FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install FFmpeg
RUN apk add --no-cache ffmpeg

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o transcode-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ffmpeg ca-certificates
WORKDIR /app
COPY --from=builder /build/transcode-service .
COPY --from=builder /build/go.mod .

EXPOSE 8011
CMD ["./transcode-service"]
