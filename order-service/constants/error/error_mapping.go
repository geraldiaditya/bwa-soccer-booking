package error

import (
	"errors"
	"net/http"
	errOrder "order-service/constants/error/order"
)

func ErrStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

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

func ErrMapping(err error) bool {
	if err == nil {
		return false
	}
	var (
		GeneralErrors = GeneralErrors
		OrderErrors   = errOrder.OrderErrors
	)

	allErrors := make([]error, 0)
	allErrors = append(allErrors, GeneralErrors...)
	allErrors = append(allErrors, OrderErrors...)

	for _, item := range allErrors {
		if err.Error() == item.Error() {
			return true
		}
	}
	return false
}
