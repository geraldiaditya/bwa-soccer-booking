package dto

import "payment-service/constants"

type PaymentHistoryRequest struct {
	PaymentId uint                          `json:"paymentID"`
	Status    constants.PaymentStatusString `json:"status"`
}
