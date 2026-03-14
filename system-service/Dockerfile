FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY system-service/ system-service/

WORKDIR /build/system-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/system-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/system-service .

EXPOSE 8010
CMD ["./system-service"]
