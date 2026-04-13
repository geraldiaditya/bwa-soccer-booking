package error

import (
	errField "field-service/constants/error/field"
	errFieldSchedule "field-service/constants/error/field_schedule"
	errTime "field-service/constants/error/time"
	"net/http"
)

func ErrStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch err.Error() {
	case errField.ErrFieldNotFound.Error(),
		errFieldSchedule.ErrFieldScheduleNotFound.Error(),
		errTime.ErrTimeNotFound.Error():
		return http.StatusNotFound
	case errFieldSchedule.ErrFieldScheduleIsExist.Error():
		return http.StatusConflict
	case ErrUnauthorized.Error():
		return http.StatusUnauthorized
	case ErrInternalServerError.Error():
		return http.StatusInternalServerError
	case ErrTooManyRequests.Error():
		return http.StatusTooManyRequests
	default:
		return http.StatusBadRequest
	}
}

func ErrMapping(err error) bool {
	if err == nil {
		return false
	}

	allErrors := make([]error, 0)
	allErrors = append(allErrors, GeneralErrors...)
	allErrors = append(allErrors, errField.FieldErrors...)
	allErrors = append(allErrors, errFieldSchedule.FieldScheduleErrors...)
	allErrors = append(allErrors, errTime.TimeErrors...)

	for _, item := range allErrors {
		if err.Error() == item.Error() {
			return true
		}
	}
	return false
}
