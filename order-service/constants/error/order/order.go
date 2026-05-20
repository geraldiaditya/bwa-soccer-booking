package error

import "errors"

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrFieldAlreadyBooked    = errors.New("field schedule already booked")
	ErrUnknownPaymentStatus  = errors.New("unknown payment status")
)

var OrderErrors = []error{
	ErrOrderNotFound,
	ErrFieldAlreadyBooked,
	ErrUnknownPaymentStatus,
}
