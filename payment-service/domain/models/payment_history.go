package models

import (
	"payment-service/constants"
	"time"
)

type PaymentHistory struct {
	ID        uint                          `gorm:"primary_key;autoIncrement"`
	PaymentID uint                          `gorm:"type:bigint;not null"`
	Status    constants.PaymentStatusString `gorm:"type:varchar(50);not null"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
