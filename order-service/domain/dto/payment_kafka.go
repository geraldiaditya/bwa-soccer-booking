package dto

import (
	"github.com/google/uuid"
	"order-service/constants"
	"time"
)

type PaymentData struct {
	OrderID   uuid.UUID                     `json:"orderId"`
	PaymentID uuid.UUID                     `json:"paymentId"`
	Status    constants.PaymentStatusString `json:"status"`
	ExpiredAt *time.Time                    `json:"expiredAt"`
	PaidAt    *time.Time                    `json:"paidAt"`
}

type PaymentContent struct {
	Event    KafkaEvent             `json:"event"`
	Metadata KafkaMetaData          `json:"metadata"`
	Body     KafkaBody[PaymentData] `json:"body"`
}
