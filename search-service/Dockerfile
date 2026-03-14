FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY search-service/ search-service/

WORKDIR /build/search-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/search-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/search-service .

EXPOSE 8016
CMD ["./search-service"]
