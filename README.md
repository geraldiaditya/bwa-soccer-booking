# BWA Soccer Booking

A microservices-based soccer field booking platform built with Go.

## Services

| Service | Port | Description |
|---|---|---|
| [user-service](./user-service) | 8001 | User management & authentication (JWT) |
| [field-service](./field-service) | 8002 | Field and schedule management |
| [order-service](./order-service) | 8003 | Order processing with Kafka consumer |
| [payment-service](./payment-service) | 8004 | Payment processing via Midtrans |

## Architecture

```
                        ┌─────────────────────────────────────────┐
                        │              Client (HTTP)               │
                        └────┬──────────┬──────────┬──────────────┘
                             │          │          │
                    ┌────────▼──┐ ┌─────▼────┐ ┌──▼────────┐
                    │   user-   │ │  field-  │ │  order-   │
                    │  service  │ │ service  │ │  service  │
                    └────────┬──┘ └─────┬────┘ └──┬──────┬─┘
                             │          │          │      │
                    ┌────────▼──────────▼──────────▼─┐   │ Kafka
                    │           PostgreSQL            │   │ event
                    └────────────────────────────────┘   │
                                                    ┌─────▼──────┐
                                                    │  payment-  │
                                                    │  service   │
                                                    └────────────┘
```

**Inter-service communication:** HTTP with HMAC-SHA256 API key signature  
**Async messaging:** Apache Kafka (order → payment events)  
**Auth:** JWT Bearer token

## Tech Stack

- **Language:** Go 1.23
- **Framework:** Gin
- **ORM:** GORM + PostgreSQL
- **Messaging:** Apache Kafka (IBM/sarama)
- **Auth:** JWT (`golang-jwt/jwt`)
- **Payment:** Midtrans
- **Docs:** Swagger (swaggo)
- **Container:** Docker + Docker Compose

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.23+

### Run all services with Docker

```bash
# Start infrastructure + all services
docker-compose up -d --build

# View logs
docker-compose logs -f user-service
```

### Run a single service locally

```bash
cd user-service          # or field-service, order-service, payment-service

cp config.json.example config.json   # fill in your DB credentials
go mod tidy

make watch-prepare       # install Air (first time only)
make watch               # run with hot reload
```

### Database migration & seeding

```bash
# Inside each service directory
./[service-name] migrate
./[service-name] seed
```

## API Documentation

Each service exposes Swagger UI when running:

| Service | Swagger URL |
|---|---|
| user-service | http://localhost:8001/swagger/index.html |
| field-service | http://localhost:8002/swagger/index.html |
| order-service | http://localhost:8003/swagger/index.html |
| payment-service | http://localhost:8004/swagger/index.html |

## Project Structure

Each service follows the same clean architecture pattern:

```
[service]/
├── cmd/              # CLI entrypoint (serve, migrate, seed)
├── config/           # App config + DB connection
├── controllers/      # HTTP handlers
├── services/         # Business logic
├── repositories/     # Data access (GORM)
├── domain/
│   ├── models/       # GORM models
│   └── dto/          # Request/response structs
├── middlewares/      # Auth, RBAC, rate limiter, error handler
├── routes/           # Route definitions
├── clients/          # HTTP clients for inter-service calls
└── docs/             # Generated Swagger docs
```

## Development Commands

```bash
make watch      # hot reload with Air
make test       # run unit tests
make swagger    # regenerate Swagger docs
make build      # build binary
make docker-compose  # run with Docker
```
