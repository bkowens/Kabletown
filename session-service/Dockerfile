FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY shared/ shared/
COPY session-service/ session-service/

WORKDIR /build/session-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/session-service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/session-service .

EXPOSE 8007
CMD ["./session-service"]
