# Build stage
FROM golang:1.23.4-alpine3.21 AS builder

RUN apk add --no-cache git openssh tzdata build-base

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN go build -ldflags="-w -s" -o field-service cmd/main.go

# Final stage
FROM alpine:3.21

RUN apk add --no-cache tzdata curl ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/field-service .
# Copy config if needed (though usually passed via env or consul)
# COPY --from=builder /app/config.json .

# Use non-root user
USER appuser

EXPOSE 8002

ENTRYPOINT ["./field-service"]