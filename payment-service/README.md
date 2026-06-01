# Payment Service

Payment Service owns payment creation, payment status reads, Midtrans webhook handling, invoice upload, and payment status publication for the BWA Soccer Booking system.

**Default port:** `8004`

## Responsibility

- Create Midtrans Snap payment links for soccer field booking orders.
- Persist payment records and payment history with GORM/PostgreSQL.
- Receive Midtrans status webhooks and update local payment status.
- Generate paid invoices for settled payments and upload the PDF to Google Cloud Storage.
- Publish payment status events to Kafka so other services can react to pending, settled, or expired payments.

## Architecture

```
HTTP clients
  |
  v
Gin routes/controllers  ->  payment service
                              |
                              +-> payment repositories -> PostgreSQL
                              +-> Midtrans Snap client
                              +-> GCS invoice storage
                              +-> Kafka producer
```

Key directories:

```
payment-service/
|-- cmd/                  # Cobra serve command entrypoint
|-- clients/              # Midtrans and internal user-service clients
|-- common/               # Shared response, utility, PDF, and GCS helpers
|-- config/               # JSON/Consul config binding and database setup
|-- controllers/http/     # Gin handlers
|-- controllers/kafka/    # Kafka producer wrapper
|-- domain/               # DTOs and GORM models
|-- middlewares/          # Panic recovery, rate limiting, auth, role checks
|-- repositories/         # Payment and payment-history data access
|-- routes/               # /api/v1/payment route registration
|-- services/             # Payment business logic
|-- templates/            # Invoice HTML template
```

## HTTP API

Routes are mounted under `/api/v1/payment`.

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/payment/webhook` | Midtrans callback payload | Update payment status from Midtrans webhook |
| `GET` | `/api/v1/payment` | JWT + service signature, admin/customer | List payments with pagination |
| `GET` | `/api/v1/payment/:uuid` | JWT + service signature, admin/customer | Get one payment by payment UUID |
| `POST` | `/api/v1/payment` | JWT + service signature, customer | Create a payment link |

There is no generated Swagger/OpenAPI bundle in the current repository. Use the route/controller DTOs as the source of truth until API docs are generated.

## Auth And Signature

Authenticated routes require:

- `Authorization: Bearer <token>`
- `x-service-name`
- `x-request-at`
- `x-apikey`

`x-apikey` is validated as:

```text
sha256("<x-service-name>:<signatureKey>:<x-request-at>")
```

After signature validation, the middleware calls user-service with the bearer token and checks the allowed role for the route. The Midtrans webhook endpoint is intentionally outside the JWT middleware because it is called by Midtrans.

## Create Payment Flow

1. A customer calls `POST /api/v1/payment` with an order ID, expiration time, amount, customer details, and item details.
2. The service rejects requests whose `expiredAt` is not in the future.
3. The Midtrans client creates a Snap payment transaction and returns a redirect URL.
4. The payment repository persists a payment with `initial` status and the redirect URL.
5. The payment-history repository records the initial status.
6. The response returns the payment UUID, order ID, amount, status, payment link, and description.

The service wraps database writes in a GORM transaction. Midtrans is called before local persistence, so a local database failure can leave a Midtrans transaction that must be reconciled operationally.

## Midtrans Webhook Flow

1. Midtrans calls `POST /api/v1/payment/webhook`.
2. The service finds the payment by `order_id`.
3. It maps `transaction_status` to the internal status enum and updates transaction ID, VA number, bank, acquirer, and paid time.
4. It writes a payment-history row for the new status.
5. For `settlement`, it generates an invoice PDF from `templates/invoice.html`, uploads it to GCS, and stores the invoice URL on the payment.
6. After the database transaction succeeds, it publishes a Kafka status event.

Supported status names in code are `pending`, `settlement`, and `expire`.

## Kafka Event Flow

Webhook processing publishes a JSON message through the configured Kafka producer. The topic comes from `config.json` at `kafka.topic`.

The event name is the uppercase transaction status (`PENDING`, `SETTLEMENT`, or `EXPIRE`). The body includes order ID, payment ID, status, paid time, and expiration time.

The current codebase only contains a Kafka producer in payment-service. It does not contain a Kafka consumer that creates payments from order events.

## Local Setup

From this directory:

```bash
cp config.json.example config.json
go mod download
```

Fill `config.json` with local PostgreSQL, Kafka, Midtrans, GCS, user-service, and signature settings. Do not commit real credentials.

Run the service:

```bash
go run .
```

Or use hot reload:

```bash
make watch-prepare
make watch
```

The Cobra command currently serves the HTTP API and runs `AutoMigrate` for payment tables during startup. Separate `migrate` and `seed` subcommands are not present.

## Build And Test

```bash
go test ./... -cover
go build ./...
```

The Makefile also provides:

```bash
make build
make docker-compose
```

## Design Decisions And Tradeoffs

- The service keeps payment state local instead of relying only on Midtrans, which gives the platform fast status reads and an audit trail.
- Payment history is append-only at the service layer so webhook status changes can be reviewed later.
- Invoice generation and upload happen inside webhook processing, which keeps the settlement flow simple but makes webhook latency depend on PDF rendering and GCS availability.
- Kafka publication happens after the database transaction succeeds, so local state is not rolled back if Kafka publishing fails. A retry/outbox mechanism would make this more reliable.
- The service currently accepts the webhook route without JWT because Midtrans cannot call with the platform's internal user token. Production deployments should validate Midtrans signatures before trusting the payload.
