FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY item-service/ item-service/

WORKDIR /build/item-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/item-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/item-service .

EXPOSE 8017
CMD ["./item-service"]
