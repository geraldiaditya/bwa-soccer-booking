package dto

import (
	"github.com/google/uuid"
	"order-service/constants"
	"time"
)

type OrderRequest struct {
	FieldScheduleIDs []string `json:"fieldScheduleIDs" validate:"required,min=1,dive,uuid4"`
}

type OrderRequestParam struct {
	Page       int     `form:"page"       validate:"required,min=1"`
	Limit      int     `form:"limit"      validate:"required,min=1,max=100"`
	SortColumn *string `form:"sortColumn" validate:"omitempty,oneof=created_at amount status date"`
	SortOrder  *string `form:"sortOrder"  validate:"omitempty,oneof=asc desc"`
}

type OrderResponse struct {
	UUID        uuid.UUID                   `json:"uuid"`
	Code        string                      `json:"code"`
	UserName    string                      `json:"userName"`
	Amount      float64                     `json:"amount"`
	Status      constants.OrderStatusString `json:"status"`
	PaymentLink string                      `json:"paymentLink"`
	OrderDate   time.Time                   `json:"orderDate"`
	CreatedAt   time.Time                   `json:"createdAt"`
	UpdateAt    time.Time                   `json:"updatedAt"`
}

type OrderByUserIDResponse struct {
	Code        string                      `json:"code"`
	Amount      string                      `json:"amount"`
	Status      constants.OrderStatusString `json:"status"`
	OrderDate   string                      `json:"orderDate"`
	PaymentLink string                      `json:"paymentLink"`
	InvoiceLink *string                     `json:"invoiceLink"`
}
