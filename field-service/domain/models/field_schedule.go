package models

import (
	"field-service/constants"
	"github.com/google/uuid"
	"time"
)

type FieldSchedule struct {
	ID        uint                          `gorm:"primaryKey;autoIncrement"`
	UUID      uuid.UUID                     `gorm:"type:uuid;uniqueIndex;not null"`
	FieldID   uint                          `gorm:"type:int;index:idx_field_schedule_lookup;not null"`
	TimeID    uint                          `gorm:"type:int;not null"`
	Date      time.Time                     `gorm:"type:date;index:idx_field_schedule_lookup;not null"`
	Status    constants.FieldScheduleStatus `gorm:"type:int;not null"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time

	Field Field `gorm:"foreignKey:field_id;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Time  Time  `gorm:"foreignKey:time_id;references:id;"`
}
