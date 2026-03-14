FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY metadata-service/ metadata-service/

WORKDIR /build/metadata-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/metadata-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/metadata-service .

EXPOSE 8008
CMD ["./metadata-service"]
