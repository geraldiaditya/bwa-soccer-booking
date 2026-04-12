package dto

import (
	"github.com/google/uuid"
	"time"
)

type PaymentRequest struct {
	PaymentLink    string         `json:"paymentLink"`
	OrderID        string         `json:"orderID"`
	ExpiredAt      time.Time      `json:"expiredAt"`
	Amount         float64        `json:"amount"`
	Description    string         `json:"description"`
	CustomerDetail CustomerDetail `json:"customerDetail"`
	ItemDetails    []ItemDetail   `json:"itemDetails"`
}

type CustomerDetail struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
type ItemDetail struct {
	ID       uuid.UUID `json:"id"`
	Amount   float64   `json:"amount"`
	Name     string    `json:"name"`
	Quantity int       `json:"quantity"`
}
