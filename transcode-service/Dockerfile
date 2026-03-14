FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ffmpeg gcc musl-dev

WORKDIR /build

COPY shared/ shared/
COPY transcode-service/ transcode-service/

WORKDIR /build/transcode-service

RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/transcode-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ffmpeg ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/transcode-service .

EXPOSE 8011
CMD ["./transcode-service"]
