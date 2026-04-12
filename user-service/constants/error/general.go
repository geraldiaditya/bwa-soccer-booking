package error

import "errors"

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrSQLError            = errors.New("database server failed to execute query")
	ErrRequestValidation   = errors.New("request validation error")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("data not found")
	ErrTooManyRequests     = errors.New("too many requests")
)

var GeneralErrors = []error{
	ErrInternalServerError,
	ErrSQLError,
	ErrRequestValidation,
	ErrUnauthorized,
	ErrForbidden,
	ErrNotFound,
	ErrTooManyRequests,
}
