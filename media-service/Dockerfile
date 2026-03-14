FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o media-service ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/media-service .

EXPOSE 8005
CMD ["./media-service"]
