package error

import "errors"

var (
	ErrFieldScheduleNotFound = errors.New("field schedule not found")
	ErrFieldScheduleIsExist  = errors.New("field schedule is exist")
)

var FieldScheduleErrors = []error{
	ErrFieldScheduleNotFound,
	ErrFieldScheduleIsExist,
}
