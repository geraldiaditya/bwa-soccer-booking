package models

import (
	"github.com/google/uuid"
	"time"
)

type OrderField struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	OrderID         uint      `gorm:"type:bigint;not null;index"`
	FieldScheduleID uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}
