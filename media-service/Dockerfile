FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY media-service/ media-service/

WORKDIR /build/media-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/media-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/media-service .

EXPOSE 8005
CMD ["./media-service"]
