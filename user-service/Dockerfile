FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY user-service/ user-service/

WORKDIR /build/user-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/user-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/user-service .

EXPOSE 8002
CMD ["./user-service"]
