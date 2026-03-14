FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY notification-service/ notification-service/

WORKDIR /build/notification-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/notification-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/notification-service .

EXPOSE 8014
CMD ["./notification-service"]
