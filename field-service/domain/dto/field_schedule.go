package dto

import (
	"field-service/constants"
	"github.com/google/uuid"
	"time"
)

type FieldScheduleRequest struct {
	FieldID string   `json:"fieldId" validate:"required"`
	Date    string   `json:"date" validate:"required,datetime=2006-01-02"`
	TimeIDs []string `json:"timeIDs" validate:"required"`
}

type GenerateFieldScheduleForOneMonthRequest struct {
	FieldID string `json:"fieldId" validate:"required"`
}

type UpdateFieldScheduleRequest struct {
	Date   string `json:"date" validate:"required"`
	TimeID string `json:"timeID" validate:"required"`
}

type UpdateStatusFieldScheduleRequest struct {
	FieldScheduleIDs []string `json:"fieldScheduleIDs" validate:"required"`
}

type FieldScheduleResponse struct {
	UUID         uuid.UUID                         `json:"uuid"`
	FieldName    string                            `json:"fieldName"`
	PricePerHour int                               `json:"pricePerHour"`
	Date         string                            `json:"date"`
	Status       constants.FieldScheduleStatusName `json:"status"`
	Time         string                            `json:"time"`
	CreatedAt    *time.Time                        `json:"createdAt"`
	UpdatedAt    *time.Time                        `json:"updatedAt"`
}

type FieldScheduleForBookingResponse struct {
	UUID         uuid.UUID                         `json:"uuid"`
	PricePerHour string                            `json:"pricePerHour"`
	Date         string                            `json:"date"`
	Status       constants.FieldScheduleStatusName `json:"status"`
	Time         string                            `json:"time"`
}

type FieldScheduleRequestParam struct {
	Page       int     `form:"page" validate:"required,min=1"`
	Limit      int     `form:"limit" validate:"required,min=1,max=100"`
	SortColumn *string `form:"sortColumn" validate:"omitempty,oneof=created_at status date"`
	SortOrder  *string `form:"sortOrder" validate:"omitempty,oneof=asc desc"`
}

type FieldScheduleByFieldIDAndDateRequestParam struct {
	Date string `form:"date" validate:"required,datetime=2006-01-02"`
}
