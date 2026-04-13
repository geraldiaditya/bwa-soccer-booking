# Field Service

Manages soccer field listings, availability schedules, and time slot configuration. Calls user-service internally to validate request signatures.

**Port:** `8002`

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/fields` | JWT (admin) | Create field |
| GET | `/api/v1/fields` | — | List all fields |
| GET | `/api/v1/fields/:uuid` | — | Get field detail |
| PUT | `/api/v1/fields/:uuid` | JWT (admin) | Update field |
| DELETE | `/api/v1/fields/:uuid` | JWT (admin) | Delete field |
| POST | `/api/v1/field-schedules` | JWT (admin) | Create schedule |
| GET | `/api/v1/field-schedules` | — | List schedules |
| GET | `/api/v1/field-schedules/:uuid` | Signature | Internal: get schedule |

## Directory Structure

```
field-service/
├── cmd/              # CLI entrypoint (serve, migrate, seed)
├── clients/          # HTTP client for user-service
├── config/           # App config + DB connection
├── controllers/      # HTTP handlers (field, field_schedule)
├── services/         # Business logic
├── repositories/     # Data access (GORM)
├── domain/
│   ├── models/       # GORM models (Field, FieldSchedule)
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
./field-service migrate
./field-service seed
```

## Build

```bash
make build
```

## API Docs

Swagger UI: http://localhost:8002/swagger/index.html
