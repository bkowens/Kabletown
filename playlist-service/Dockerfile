FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY playlist-service/ playlist-service/

WORKDIR /build/playlist-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/playlist-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/playlist-service .

EXPOSE 8013
CMD ["./playlist-service"]
