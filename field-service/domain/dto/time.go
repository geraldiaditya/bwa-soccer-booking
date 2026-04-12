package dto

import (
	"github.com/google/uuid"
	"time"
)

type TimeRequest struct {
	StartTime string `form:"startTime" validate:"required"`
	EndTime   string `form:"endTime" validate:"required"`
}

type TimeResponse struct {
	UUID      uuid.UUID  `json:"uuid"`
	StartTime string     `json:"startTime"`
	EndTime   string     `json:"endTime"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}
