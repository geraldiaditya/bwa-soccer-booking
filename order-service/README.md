# Order Service

Processes field booking orders. Calls user-service and field-service to validate bookings, then publishes order events to Kafka. Consumes payment callback events from Kafka to update order status.

**Port:** `8003`

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/orders` | JWT | Create order |
| GET | `/api/v1/orders` | JWT | List my orders |
| GET | `/api/v1/orders/:uuid` | JWT | Get order detail |

## Event Flow

```
order-service  ──[order.created]──►  Kafka  ──►  payment-service
payment-service ──[payment.callback]──►  Kafka  ──►  order-service
```

## Directory Structure

```
order-service/
├── cmd/              # CLI entrypoint (serve, migrate, seed)
├── clients/          # HTTP clients for user-service & field-service
├── config/           # App config + DB + Kafka connection
├── controllers/
│   ├── http/         # HTTP handlers
│   └── kafka/        # Kafka consumer handlers
├── services/         # Business logic
├── repositories/     # Data access (GORM)
├── domain/
│   ├── models/       # GORM models (Order)
│   └── dto/          # Request/response structs
├── middlewares/      # JWT auth, RBAC, HMAC signature
├── routes/           # Route definitions
└── docs/             # Generated Swagger docs
```

## Setup

```bash
cp config.json.example config.json   # fill in DB credentials & service keys
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
./order-service migrate
./order-service seed
```

## Build

```bash
make build
```

## API Docs

Swagger UI: http://localhost:8003/swagger/index.html
