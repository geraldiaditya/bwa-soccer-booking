# Field Service

Field Service owns soccer field inventory, bookable schedules, and time slot configuration for the BWA Soccer Booking system. It exposes public read endpoints for customers, admin-only write endpoints for back-office operations, and signed internal endpoints used by other services.

Default port: `8002`

## Responsibilities

- Manage field records, including code, name, hourly price, and uploaded images.
- Manage reusable time slots.
- Generate and maintain field availability schedules.
- Serve customer-facing schedule data for booking flows.
- Validate JWT-bearing admin/customer requests through user-service.
- Validate signed internal requests with `x-service-name`, `x-request-at`, and `x-apikey` headers.

## Architecture

```
HTTP request
  -> Gin route
  -> global middleware: logging, panic recovery, security headers, CORS, rate limiting
  -> auth middleware: API signature and optional JWT role check
  -> controller: bind and validate request data
  -> service: business rules and DTO mapping
  -> repository: GORM/PostgreSQL persistence
```

The code follows a simple layered structure:

- `cmd/` starts the Cobra command and Swagger metadata.
- `internal/app/` wires configuration, database, routes, middleware, Swagger, and graceful shutdown.
- `routes/` defines URL groups and route-level authentication.
- `controllers/` handles HTTP binding, validation, and response mapping.
- `services/` contains business rules for fields, schedules, and time slots.
- `repositories/` isolates GORM queries and persistence.
- `domain/models/` stores GORM models.
- `domain/dto/` stores request and response contracts.
- `clients/` contains internal service clients, currently user-service.
- `middlewares/` contains request logging, auth, RBAC, CORS, security headers, rate limiting, and panic handling.
- `docs/` contains generated Swagger files.

## Request Flow

Field and schedule writes are protected by both request signature validation and role checks. A typical admin write request flows through:

1. `Authenticate()` validates the shared service signature and extracts the bearer token.
2. `CheckRole()` calls user-service with the token and verifies the role.
3. The controller binds JSON, form, path, or query data and validates it.
4. The service enforces business rules, such as duplicate schedule checks.
5. The repository performs GORM operations against PostgreSQL.

Public read-like endpoints still use `AuthenticateWithoutToken()`, so callers need a valid internal signature even when no JWT is required.

## Authentication And Signature

The service uses two layers of request protection:

- API signature: `x-apikey = sha256(x-service-name + ":" + signatureKey + ":" + x-request-at)`.
- JWT role check: `Authorization: Bearer <token>`, forwarded to user-service for user and role lookup.

Admin-only endpoints require the `Admin` role. Some paginated reads allow both `Admin` and `Customer`.

## Main Routes

| Method | Path | Access | Description |
| --- | --- | --- | --- |
| GET | `/` | public | Service welcome response |
| GET | `/health` | public | Health check |
| GET | `/api/v1/field` | signed request | List field summaries |
| GET | `/api/v1/field/:uuid` | signed request | Get field detail |
| GET | `/api/v1/field/pagination` | signed request + Admin/Customer JWT | Paginated fields |
| POST | `/api/v1/field` | signed request + Admin JWT | Create field |
| PUT | `/api/v1/field/:uuid` | signed request + Admin JWT | Update field |
| DELETE | `/api/v1/field/:uuid` | signed request + Admin JWT | Delete field |
| GET | `/api/v1/field/schedule/lists/:uuid` | signed request | List schedules for a field and date |
| PATCH | `/api/v1/field/schedule/status` | signed internal request | Mark schedules as booked from a booking/order flow |
| GET | `/api/v1/field/schedule/pagination` | signed request + Admin/Customer JWT | Paginated schedules |
| GET | `/api/v1/field/schedule/:uuid` | signed request + Admin/Customer JWT | Get schedule detail |
| POST | `/api/v1/field/schedule` | signed request + Admin JWT | Create schedules for selected time slots |
| POST | `/api/v1/field/schedule/one-month` | signed request + Admin JWT | Generate one month of schedules |
| PUT | `/api/v1/field/schedule/:uuid` | signed request + Admin JWT | Update schedule date/time |
| DELETE | `/api/v1/field/schedule/:uuid` | signed request + Admin JWT | Delete schedule |
| GET | `/api/v1/time` | signed request + Admin JWT | List time slots |
| GET | `/api/v1/time/:uuid` | signed request + Admin JWT | Get time slot detail |
| POST | `/api/v1/time` | signed request + Admin JWT | Create time slot |

## Local Setup

Prerequisites:

- Go 1.23+
- PostgreSQL
- Optional: Docker and Docker Compose
- Optional: `swag` CLI when regenerating Swagger docs

Create local config:

```bash
cp config.json.example config.json
```

Fill in database credentials, `signatureKey`, user-service host/signature key, rate limit settings, and optional GCS settings. When `gcsPrivateKey` is empty, the service falls back to the local GCS client setup used by the code.

Install dependencies:

```bash
go mod tidy
```

Run with hot reload:

```bash
make watch-prepare
make watch
```

Run without hot reload:

```bash
go run main.go
```

Build:

```bash
make build
```

Docker:

```bash
docker-compose up -d --build
```

## Database

The application runs `AutoMigrate` during startup for:

- `fields`
- `field_schedules`
- `times`

The schedule model indexes `field_id` and `date` for lookup-heavy booking flows.

There is no active seed command in the current Cobra entrypoint. For local development, create time slots through the `POST /api/v1/time` admin endpoint or insert controlled seed rows with your normal database tooling before creating schedules.

Booking schedule dates are returned in short Indonesian display format, for example `22 Mei`.

## Testing

Run all tests with coverage:

```bash
go test ./... -cover
```

Or use the Make target:

```bash
make test
```

The test suite focuses on service business rules, repository pagination helpers, DTO mappers, and utility behavior.

## API Docs

Swagger UI is available after the service starts:

```
http://localhost:8002/swagger/index.html
```

Regenerate docs after changing controller annotations:

```bash
make swagger
```

## Design Decisions And Tradeoffs

- Layered packages keep HTTP, business rules, and persistence concerns separate enough for focused tests and recruiter review.
- GORM keeps data access concise while repository interfaces keep service tests independent from PostgreSQL.
- Schedule creation checks existing schedules before insert to preserve the current duplicate-prevention behavior.
- The service uses shared API signatures for service-to-service trust and delegates JWT role validation to user-service, avoiding duplicated user authorization logic.
- Global middleware centralizes operational concerns: request IDs, logging, recovery, security headers, CORS, and rate limiting.
- Startup performs migrations automatically, which is convenient for portfolio and local development but should be reviewed before production environments that require controlled migrations.
