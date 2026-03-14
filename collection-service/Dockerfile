FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY collection-service/ collection-service/

WORKDIR /build/collection-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/collection-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/collection-service .

EXPOSE 8012
CMD ["./collection-service"]
