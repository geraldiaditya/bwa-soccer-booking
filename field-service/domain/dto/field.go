package dto

import (
	"github.com/google/uuid"
	"mime/multipart"
	"time"
)

type FieldRequest struct {
	Name         string                 `form:"name" validate:"required,min=3,max=100"`
	Code         string                 `form:"code" validate:"required,max=15"`
	PricePerHour int                    `form:"pricePerHour" validate:"required,gt=0"`
	Images       []multipart.FileHeader `form:"images" validate:"required"`
}

type UpdateFieldRequest struct {
	Name         string                 `form:"name" validate:"required,min=3,max=100"`
	Code         string                 `form:"code" validate:"required,max=15"`
	PricePerHour int                    `form:"pricePerHour" validate:"required,gt=0"`
	Images       []multipart.FileHeader `form:"images"`
}

type FieldResponse struct {
	UUID         uuid.UUID  `json:"uuid"`
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	PricePerHour int        `json:"pricePerHour"`
	Images       []string   `json:"images"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

type FieldDetailResponse struct {
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	PricePerHour int        `json:"pricePerHour"`
	Images       []string   `json:"images"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

type FieldRequestParam struct {
	Page       int     `form:"page" validate:"required,min=1"`
	Limit      int     `form:"limit" validate:"required,min=1,max=100"`
	SortColumn *string `form:"sortColumn" validate:"omitempty,oneof=created_at price_per_hour name"`
	SortOrder  *string `form:"sortOrder" validate:"omitempty,oneof=asc desc"`
}
