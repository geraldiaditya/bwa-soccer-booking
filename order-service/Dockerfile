FROM golang:1.23.4-alpine3.21 AS builder

RUN apk add --no-cache git tzdata build-base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o order-service ./main.go

FROM alpine:3.21

RUN apk add --no-cache tzdata wget ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/order-service .

USER appuser

EXPOSE 8003

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8003/health || exit 1

ENTRYPOINT ["/app/order-service"]