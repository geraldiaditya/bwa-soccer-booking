package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	UUID        uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null"`
	Name        string     `gorm:"type:varchar(100);not null"`
	Username    string     `gorm:"type:varchar(20);uniqueIndex;not null"`
	Password    string     `gorm:"type:varchar(255);not null"`
	PhoneNumber string     `gorm:"type:varchar(15);not null"`
	Email       string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	RoleID      uint       `gorm:"type:uint;not null"`
	CreateAt    *time.Time `gorm:"autoCreateTime"`
	UpdateAt    *time.Time `gorm:"autoUpdateTime"`
	Role        Role       `gorm:"foreignKey:role_id;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
