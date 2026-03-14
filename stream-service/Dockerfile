FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY stream-service/ stream-service/

WORKDIR /build/stream-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/stream-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/stream-service .

EXPOSE 8006
CMD ["./stream-service"]
