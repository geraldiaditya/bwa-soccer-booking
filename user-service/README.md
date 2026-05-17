# User Service

Handles user registration, authentication, and profile management. Issues JWT tokens consumed by all other services. Active JWT sessions are cached in Redis so tokens can be revoked before their expiry.

**Port:** `8001`

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/users/register` | — | Register new user |
| POST | `/api/v1/users/login` | — | Login, returns JWT |
| GET | `/api/v1/users/me` | JWT | Get logged-in user |
| GET | `/api/v1/users/:uuid` | Signature | Internal: get user by UUID |

## Directory Structure

```
user-service/
├── cmd/              # CLI entrypoint (serve, migrate, seed)
├── config/           # App config + DB connection
├── controllers/      # HTTP handlers
├── services/         # Business logic
├── repositories/     # Data access (GORM)
├── domain/
│   ├── models/       # GORM models (User, Role)
│   └── dto/          # Request/response structs
├── middlewares/      # JWT auth, RBAC, rate limiter
├── routes/           # Route definitions
└── docs/             # Generated Swagger docs
```

## Setup

```bash
cp config.json.example config.json   # fill in DB credentials, Redis, and JWT secret
go mod tidy
```

For Docker Compose, set `redis.host` to `redis`. For local development outside Docker, use `localhost` when Redis is exposed on port `6379`.

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
./user-service migrate
./user-service seed
```

## Build

```bash
make build
```

## API Docs

Swagger UI: http://localhost:8001/swagger/index.html

## Test

```bash
make test
```
