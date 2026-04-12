package clients

import "github.com/google/uuid"

type FieldResponse struct {
	Code    int       `json:"code"`
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Data    FieldData `json:"data"`
}

type FieldData struct {
	UUID          uuid.UUID `json:"uuid"`
	FieldName     string    `json:"fieldName"`
	PricePerHours float64   `json:"pricePerHour"`
	Date          string    `json:"date"`
	StartTime     string    `json:"startTime"`
	EndTime       string    `json:"endTime"`
	Status        string    `json:"status"`
	CreatedAt     string    `json:"createdAt"`
	UpdatedAt     string    `json:"updatedAt"`
}
