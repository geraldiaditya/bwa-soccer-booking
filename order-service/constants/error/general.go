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
	ErrInvalidUploadFile   = errors.New("invalid upload file")
	ErrSizeTooBig          = errors.New("file size too big")
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
