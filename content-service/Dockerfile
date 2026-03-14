FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY content-service/ content-service/

WORKDIR /build/content-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/content-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/content-service .

EXPOSE 8009
CMD ["./content-service"]
