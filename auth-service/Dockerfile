FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY auth-service/ auth-service/

WORKDIR /build/auth-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auth-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/auth-service .

EXPOSE 8001
CMD ["./auth-service"]
