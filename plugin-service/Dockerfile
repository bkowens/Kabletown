FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY plugin-service/ plugin-service/

WORKDIR /build/plugin-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/plugin-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/plugin-service .

EXPOSE 8015
CMD ["./plugin-service"]
