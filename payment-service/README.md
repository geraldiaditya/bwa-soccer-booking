# Payment Service

Handles payment processing via Midtrans. Consumes order events from Kafka to create payment transactions, serves Midtrans webhook callbacks, and publishes payment status events back to Kafka.

**Port:** `8004`

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/payments/callback` | Midtrans | Midtrans webhook callback |
| GET | `/api/v1/payments/:order_uuid` | JWT | Get payment status |

## Event Flow

```
order-service  ──[order.created]──►  Kafka  ──►  payment-service  ──►  Midtrans
Midtrans  ──►  payment-service  ──[payment.callback]──►  Kafka  ──►  order-service
```

## Directory Structure

```
payment-service/
├── cmd/              # CLI entrypoint (serve, migrate, seed)
├── clients/          # HTTP client for user-service
├── config/           # App config + DB + Kafka + Midtrans config
├── controllers/
│   ├── http/         # HTTP handlers
│   └── kafka/        # Kafka consumer handlers
├── services/         # Business logic
├── repositories/     # Data access (GORM)
├── domain/
│   ├── models/       # GORM models (Payment)
│   └── dto/          # Request/response structs
├── middlewares/      # JWT auth, RBAC, HMAC signature
├── routes/           # Route definitions
├── templates/        # HTML email/notification templates
└── docs/             # Generated Swagger docs
```

## Setup

```bash
cp config.json.example config.json   # fill in DB, Kafka, and Midtrans credentials
go mod tidy
```

## Run

```bash
make watch-prepare   # install Air (first time only)
make watch           # run with hot reload
```

## Docker

```bash
docker-compose up -d --build
```

## Database

```bash
./payment-service migrate
./payment-service seed
```

## Build

```bash
make build
```

## API Docs

Swagger UI: http://localhost:8004/swagger/index.html
