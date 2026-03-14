FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY library-service/ library-service/

WORKDIR /build/library-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/library-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/library-service .

EXPOSE 8003
CMD ["./library-service"]
