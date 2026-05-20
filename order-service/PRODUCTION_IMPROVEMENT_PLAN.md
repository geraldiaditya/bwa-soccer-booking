# order-service — Production Improvement Plan

> **Agent brief:** This document is a self-contained task list for improving `order-service` to production quality.
> Work through each tier in order. Every item lists the exact file, current code, and what to change.
> The service is a Go microservice (Gin, GORM, Kafka/IBM sarama, PostgreSQL).
> Do NOT add `Co-Authored-By: Claude` to any git commits.

---

## Tier 1 — CRITICAL (bugs that break the service at runtime)

### 1. Wrong service URLs in client registry
**File:** `clients/registry.go` lines 33–34, 41–42

Both `GetPayment()` and `GetField()` hardcode the **User** service host and signature key.
Every call to the payment service and field service is actually routed to the user service → all
cross-service calls silently fail or return wrong data.

```go
// CURRENT (WRONG):
func (registry *ClientRegistry) GetPayment() paymentClient.IPaymentClient {
    return paymentClient.NewPaymentClient(
        config.NewClientConfig(
            config.WithBaseUrl(configApp.Config.InternalService.User.Host),      // BUG
            config.WithSignatureKey(configApp.Config.InternalService.User.SignatureKey), // BUG
        ))
}
func (registry *ClientRegistry) GetField() fieldClient.IFieldClient {
    return fieldClient.NewFieldClient(
        config.NewClientConfig(
            config.WithBaseUrl(configApp.Config.InternalService.User.Host),      // BUG
            config.WithSignatureKey(configApp.Config.InternalService.User.SignatureKey), // BUG
        ))
}
```

**Fix:** Use the correct config keys:
```go
func (registry *ClientRegistry) GetPayment() paymentClient.IPaymentClient {
    return paymentClient.NewPaymentClient(
        config.NewClientConfig(
            config.WithBaseUrl(configApp.Config.InternalService.Payment.Host),
            config.WithSignatureKey(configApp.Config.InternalService.Payment.SignatureKey),
        ))
}
func (registry *ClientRegistry) GetField() fieldClient.IFieldClient {
    return fieldClient.NewFieldClient(
        config.NewClientConfig(
            config.WithBaseUrl(configApp.Config.InternalService.Field.Host),
            config.WithSignatureKey(configApp.Config.InternalService.Field.SignatureKey),
        ))
}
```

---

### 2. Wrong JSON tag on Payment config struct (config never loads)
**File:** `config/config.go` line 52

The `InternalService.Payment` field uses JSON tag `"order"` — this means the payment service config
is never populated when reading `config.json`, so `Payment.Host` is always `""`.

```go
// CURRENT (WRONG):
type InternalService struct {
    User    User    `json:"user"`
    Field   Field   `json:"field"`
    Payment Payment `json:"order"` // BUG — should be "payment"
}
```

**Fix:**
```go
type InternalService struct {
    User    User    `json:"user"`
    Field   Field   `json:"field"`
    Payment Payment `json:"payment"`
}
```

---

### 3. Kafka consumer poison pill crashes the entire partition consumer
**File:** `controllers/kafka/config/consumer_group.go` lines 53–57

After max retries, `break` exits the `for message := range messages` loop. This permanently
halts consumption on that partition — the service stops processing all future messages.
Additionally, `MarkMessage` on line 57 is unreachable after the `break`.

```go
// CURRENT (WRONG):
if err != nil {
    logrus.Errorf("failed to handle message %s with error: %s", message.Topic, err)
    break  // kills entire partition consumer loop
}
session.MarkMessage(message, time.Now().UTC().String())  // unreachable
```

**Fix:** On poison pill, log and skip — always commit the offset so Kafka doesn't redeliver:
```go
if err != nil {
    logrus.Errorf("poison pill on topic %s, skipping: %s", message.Topic, err)
    session.MarkMessage(message, "")
    continue
}
session.MarkMessage(message, "")
```

Also: `MarkMessage` metadata must be `""` (empty string), not `time.Now().UTC().String()`.
Remove the `"time"` import if it becomes unused after this change.

---

### 4. Order code format bug — first order gets space-padded code
**File:** `repositories/order/order.go` line 101

The `else` branch (first order ever) uses `%5d` (space-padded) instead of `%05d` (zero-padded).
This produces `"ORD-    1-20250902"` instead of `"ORD-00001-20250902"`, breaking the `orderCode[4:9]`
slice used for increment on subsequent orders (line 97).

```go
// CURRENT (WRONG):
} else {
    result = fmt.Sprintf("ORD-%5d-%s", 1, today)
}
```

**Fix:**
```go
} else {
    result = fmt.Sprintf("ORD-%05d-%s", 1, today)
}
```

---

### 5. `logrus.Error` used with format directive (vet error)
**File:** `common/error/error.go` line 63

`logrus.Error` does not accept format directives — `%v` is printed literally, hiding the actual error.

```go
// CURRENT (WRONG):
logrus.Error("error %v", err)
```

**Fix:**
```go
logrus.Errorf("error: %v", err)
```

---

## Tier 2 — HIGH (correctness / security issues)

### 6. Unsafe type assertions panic on nil context value
**File:** `services/order/order.go` lines 105, 133

Both `GetOrderByUserID` and `Create` do a bare type assertion on the context value without checking
for nil. If the middleware didn't set the value (e.g. misconfiguration), the service panics.

```go
// CURRENT (unsafe):
user = ctx.Value(constants.User).(*clientUser.UserData)
```

**Fix:** use the comma-ok form and return a proper error:
```go
userVal := ctx.Value(constants.User)
if userVal == nil {
    return nil, errConst.ErrUnauthorized
}
user, ok := userVal.(*clientUser.UserData)
if !ok {
    return nil, errConst.ErrUnauthorized
}
```
Apply this pattern in both `GetOrderByUserID` (line 105) and `Create` (line 133).
Add `errConst "order-service/constants/error"` import if not already present.

---

### 7. SQL injection via unsanitized SortColumn / SortOrder
**File:** `repositories/order/order.go` lines 46–48

User-supplied `SortColumn` and `SortOrder` are interpolated directly into the ORDER BY clause
without any whitelist check.

```go
// CURRENT (unsafe):
sort = fmt.Sprintf("%s %s", *param.SortColumn, *param.SortOrder)
```

**Fix:** Add a whitelist before building the sort string:
```go
var allowedColumns = map[string]bool{
    "created_at": true, "amount": true, "status": true, "date": true,
}
var allowedOrders = map[string]bool{"asc": true, "desc": true}

if param.SortColumn != nil {
    col := strings.ToLower(*param.SortColumn)
    ord := strings.ToLower(*param.SortOrder)
    if !allowedColumns[col] || !allowedOrders[ord] {
        return nil, 0, errWrap.WrapError(errConst.ErrRequestValidation)
    }
    sort = fmt.Sprintf("%s %s", col, ord)
}
```
Add `"strings"` import.

---

### 8. HTTP server has no graceful shutdown
**File:** `cmd/main.go` lines 106–110

The HTTP server runs in a bare goroutine with no shutdown handling — `SIGTERM` only triggers
Kafka consumer shutdown. In-flight HTTP requests are dropped immediately on process exit.

```go
// CURRENT:
go func() {
    port := fmt.Sprintf(":%d", config.Config.Port)
    router.Run(port)
}()
```

**Fix:** Replace `router.Run()` with `http.Server` and integrate with the signal channel:
```go
srv := &http.Server{
    Addr:    fmt.Sprintf(":%d", config.Config.Port),
    Handler: router,
}
go func() {
    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        logrus.Fatalf("HTTP server error: %v", err)
    }
}()
// Return srv from serveHttp so main can call srv.Shutdown(ctx) on SIGTERM.
```

Refactor `serveHttp` to return `*http.Server`. In `serveKafkaConsumer` (or a unified signal handler
in `Run()`), call `srv.Shutdown(shutdownCtx)` after the signal is received with a 30s timeout.
Also close the DB connection on shutdown:
```go
sqlDB, _ := db.DB()
defer sqlDB.Close()
```

---

### 9. `panic(err)` on Kafka consume error
**File:** `cmd/main.go` line 150

A transient Kafka broker error causes the whole process to crash. The consumer group library
returns errors on rebalance or temporary disconnection — these should be logged and retried.

```go
// CURRENT:
if err != nil {
    logrus.Errorf("failed to consume: %v", err)
    panic(err)
}
```

**Fix:** Log and continue — the `for` loop already retries:
```go
if err != nil {
    logrus.Errorf("failed to consume: %v", err)
    // do not panic; sarama consumer group recovers on next iteration
}
```

---

### 10. Missing database indexes on Order model
**File:** `domain/models/order.go`

UUID lookups and UserID queries run full table scans. The `OrderField` model also lacks an index
on `OrderID`.

```go
// CURRENT:
UUID   uuid.UUID `gorm:"type:uuid;not null"`
UserID uuid.UUID `gorm:"type:uuid;not null"`
```

**Fix:**
```go
// domain/models/order.go
UUID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
UserID uuid.UUID `gorm:"type:uuid;not null;index"`

// domain/models/order_field.go
OrderID         uint      `gorm:"type:bigint;not null;index"`
FieldScheduleID uuid.UUID `gorm:"type:uuid;not null;index"`
```

---

### 11. AutoCommit enabled alongside manual MarkMessage (conflicting offset strategy)
**File:** `cmd/main.go` lines 118–119

`AutoCommit.Enable = true` means Kafka commits offsets automatically every second regardless
of whether the message was processed successfully. This defeats the purpose of `MarkMessage`
and can cause messages to be silently lost after a crash.

**Fix:** Disable auto-commit and rely on manual `MarkMessage`:
```go
kafkaConsumerConfig.Consumer.Offsets.AutoCommit.Enable = false
```

---

## Tier 3 — MEDIUM (robustness / maintainability)

### 12. Config validation at startup
**File:** `config/config.go`

There is no validation of required config fields. The service starts up silently with empty
`Port`, `SignatureKey`, or service hosts, causing obscure failures later.

**Fix:** Add a `Validate()` method and call it after `Init()` in `cmd/main.go`:
```go
func (c *AppConfig) Validate() error {
    if c.Port == 0 {
        return errors.New("config: port is required")
    }
    if c.AppName == "" {
        return errors.New("config: appName is required")
    }
    if c.SignatureKey == "" {
        return errors.New("config: signatureKey is required")
    }
    if c.InternalService.User.Host == "" {
        return errors.New("config: internalService.user.host is required")
    }
    if c.InternalService.Field.Host == "" {
        return errors.New("config: internalService.field.host is required")
    }
    if c.InternalService.Payment.Host == "" {
        return errors.New("config: internalService.payment.host is required")
    }
    if len(c.Kafka.Brokers) == 0 {
        return errors.New("config: kafka.brokers is required")
    }
    return nil
}
```

In `cmd/main.go` Run:
```go
config.Init()
if err := config.Config.Validate(); err != nil {
    logrus.Fatalf("invalid config: %v", err)
}
```

---

### 13. Inline CORS middleware — extract to middlewares package
**File:** `cmd/main.go` lines 83–92

CORS is implemented as an anonymous inline function. It hardcodes `"*"` for all origins and
is not reusable.

**Fix:** Move to `middlewares/middleware.go` as a proper `CORS()` function with configurable
origins from `config.Config.AllowedOrigins []string`. Add `AllowedOrigins []string` to
`AppConfig` in `config/config.go`. In `cmd/main.go`, replace the anonymous function with:
```go
router.Use(middlewares.CORS())
```

---

### 14. Add RequestLogger and SecurityHeaders middleware
**File:** `middlewares/middleware.go`

The service logs nothing about incoming HTTP requests and sends no security headers.

**Fix:** Add two middleware functions (mirror the field-service pattern):

```go
// RequestLogger logs method, path, latency, status, and a UUID request_id.
// It also stores request_id in gin.Context for use in error responses.
func RequestLogger() gin.HandlerFunc { ... }

// SecurityHeaders adds X-Content-Type-Options, X-Frame-Options, X-XSS-Protection.
func SecurityHeaders() gin.HandlerFunc { ... }
```

Register both in `cmd/main.go` before the rate limiter:
```go
router.Use(middlewares.RequestLogger())
router.Use(middlewares.SecurityHeaders())
```

---

### 15. Error responses always use `http.StatusBadRequest`
**File:** `controllers/http/order/order.go`

All error paths return `http.StatusBadRequest` (400) regardless of the actual error type. Not-found
errors should be 404; unauthorized errors should be 401.

**Fix:** Replace `ErrMapping(err) bool` with `ErrStatusCode(err) int` in
`constants/error/error_mapping.go` (same pattern as the improved field-service):

```go
func ErrStatusCode(err error) int {
    switch {
    case errors.Is(err, ErrNotFound), errors.Is(err, errOrder.ErrOrderNotFound):
        return http.StatusNotFound
    case errors.Is(err, ErrUnauthorized):
        return http.StatusUnauthorized
    case errors.Is(err, ErrForbidden):
        return http.StatusForbidden
    case errors.Is(err, ErrInternalServerError), errors.Is(err, ErrSQLError):
        return http.StatusInternalServerError
    default:
        return http.StatusBadRequest
    }
}
```

Update `common/response/response.go` to call `ErrStatusCode` and update all controllers
to pass `errConstant.ErrStatusCode(err)` instead of a hardcoded status.

---

### 16. `field` variable in Create captures last loop iteration only
**File:** `services/order/order.go` line 182

Inside the transaction closure, `description` uses `field.FieldName`, but `field` is the outer
variable shadowed inside the loop (`:=` on line 142). After the loop exits, `field` holds the
value from the last iteration — not necessarily the right one for the payment description.

```go
// The outer `field` is only set if the inner loop uses `=` not `:=`
// But line 142 uses `:=` so outer `field` is never assigned
field, err := o.client.GetField().GetFieldByUUID(ctx, uuidParsed)  // local var, shadows outer
```

**Fix:** Accumulate field names during the pre-loop and join them for the description, or change
the loop variable to assign to the outer `field`:
```go
field, err = o.client.GetField().GetFieldByUUID(ctx, uuidParsed)  // use = not :=
```
(Remove the `err` declaration from the `var` block at the top and let it be declared by
the loop, or declare `field` at the outer scope and use `=`.)

---

### 17. Add input validation tags to DTOs
**File:** `domain/dto/order.go`

`FieldScheduleIDs` only checks `required` — no UUID format validation on individual elements.
`OrderRequestParam` has no bounds on Page and Limit.

**Fix:**
```go
type OrderRequest struct {
    FieldScheduleIDs []string `json:"fieldScheduleIDs" validate:"required,min=1,dive,uuid4"`
}

type OrderRequestParam struct {
    Page       int     `form:"page"       validate:"required,min=1"`
    Limit      int     `form:"limit"      validate:"required,min=1,max=100"`
    SortColumn *string `form:"sortColumn" validate:"omitempty,oneof=created_at amount status date"`
    SortOrder  *string `form:"sortOrder"  validate:"omitempty,oneof=asc desc"`
}
```

---

## Tier 4 — LOW (polish / observability)

### 18. Improve Dockerfile
**File:** `Dockerfile`

Current issues:
- `FROM alpine:latest` — unpinned tag; build can break or be inconsistent
- Copies entire `/app` directory from builder (includes source code, `.git`, etc.)
- No `HEALTHCHECK`
- No non-root user

**Fix:**
```dockerfile
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
```

---

### 19. Add `.dockerignore`
**File:** `.dockerignore` (new file)

Without it, `COPY . .` sends `.git/`, IDE files, test data, and secrets to the Docker build context.

```
.git
.gitignore
.idea
*.md
bin/
*_test.go
config.json
*.env
```

---

### 20. Add `/health` and `/ready` endpoints
**File:** new `controllers/health/health.go` + `routes/health/health.go`

Container orchestrators need health probes. Register before the rate limiter so probes are
never throttled.

```go
// controllers/health/health.go
func (h *HealthController) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
func (h *HealthController) Ready(c *gin.Context) {
    db, _ := h.db.DB()
    if err := db.Ping(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
```

Pass `*gorm.DB` into `serveHttp` and register health routes before the rate limiter middleware.

---

### 21. Add Swagger annotations
**File:** `cmd/main.go`, `controllers/http/order/order.go`

The service has no API documentation. Add swaggo to `go.mod` and annotate all HTTP handlers.

Add to `go.mod`:
```
github.com/swaggo/swag v1.16.1
github.com/swaggo/gin-swagger v1.6.0
github.com/swaggo/files v1.0.1
```

Add `@title`, `@version`, `@host` comments to `cmd/main.go` and `@Summary`, `@Param`, `@Success`,
`@Failure`, `@Router` comments to each controller method. Run `swag init --generalInfo cmd/main.go`.

---

### 22. Add unit tests for order service
**Files:** new `services/order/order_test.go`, `services/order/mock_test.go`

There are no tests. Create mock implementations of `IOrderRepository`, `IClientRegistry`,
`IOrderService` using `testify/mock`. Cover at minimum:
- `GetAllWithPagination` — happy path, DB error
- `Create` — happy path, field already booked, payment client error
- `HandlePayment` — settlement and expired status transitions

---

### 23. Redundant `logrus.Errorf` duplicate after retry loop
**File:** `controllers/kafka/config/consumer_group.go` lines 53–55

After the retry loop already logs each failed attempt, there is a redundant outer error log on
line 54 that fires before the `break`.

```go
// CURRENT:
if err != nil {
    logrus.Errorf("failed to handle message %s with error: %s", message.Topic, err)
    break
}
```

Once the `break → continue` fix from Tier 1 item 3 is applied, this redundant log disappears.
No separate action needed — covered by that fix.

---

## Summary checklist

| # | File | Change | Tier |
|---|------|--------|------|
| 1 | `clients/registry.go` | Fix Payment and Field client URLs | CRITICAL |
| 2 | `config/config.go` | Fix `Payment` JSON tag from `"order"` to `"payment"` | CRITICAL |
| 3 | `controllers/kafka/config/consumer_group.go` | Replace `break` with `continue` on poison pill | CRITICAL |
| 4 | `repositories/order/order.go:101` | Fix `%5d` → `%05d` in order code | CRITICAL |
| 5 | `common/error/error.go:63` | Fix `logrus.Error` → `logrus.Errorf` | CRITICAL |
| 6 | `services/order/order.go:105,133` | Safe type assertions for `constants.User` ctx value | HIGH |
| 7 | `repositories/order/order.go:46` | Whitelist SortColumn/SortOrder against SQL injection | HIGH |
| 8 | `cmd/main.go` | Graceful HTTP shutdown via `http.Server` | HIGH |
| 9 | `cmd/main.go:150` | Remove `panic(err)` in Kafka consumer loop | HIGH |
| 10 | `domain/models/order.go`, `order_field.go` | Add DB indexes (uniqueIndex UUID, index UserID) | HIGH |
| 11 | `cmd/main.go:118` | Disable Kafka AutoCommit (conflicts with MarkMessage) | HIGH |
| 12 | `config/config.go` | Add `Validate()` for required fields | MEDIUM |
| 13 | `cmd/main.go` + `middlewares/` | Extract inline CORS to `middlewares.CORS()` | MEDIUM |
| 14 | `middlewares/middleware.go` | Add `RequestLogger()` and `SecurityHeaders()` | MEDIUM |
| 15 | `constants/error/error_mapping.go` | Replace `ErrMapping(bool)` with `ErrStatusCode(int)` | MEDIUM |
| 16 | `controllers/http/order/order.go` | Use `ErrStatusCode(err)` instead of hardcoded 400 | MEDIUM |
| 17 | `services/order/order.go:142` | Fix field variable shadowing in Create loop | MEDIUM |
| 18 | `domain/dto/order.go` | Add UUID validation tags and bounds on pagination params | MEDIUM |
| 19 | `Dockerfile` | Multi-stage, non-root user, pinned alpine, HEALTHCHECK | LOW |
| 20 | `.dockerignore` | New file | LOW |
| 21 | `controllers/health/` + `routes/health/` | Health and ready endpoints | LOW |
| 22 | Controllers + `go.mod` | Swagger annotations and `swag init` | LOW |
| 23 | `services/order/order_test.go` | Unit tests with testify mocks | LOW |
