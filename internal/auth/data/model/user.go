package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	UserID      string         `gorm:"type:varchar(64);primaryKey" json:"user_id"`
	DisplayName string         `gorm:"type:varchar(128);not null" json:"display_name"`
	Status      string         `gorm:"type:varchar(32);not null;default:'ACTIVE';index" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy   string         `gorm:"type:varchar(64)" json:"created_by"`
	UpdatedBy   string         `gorm:"type:varchar(64)" json:"updated_by"`
}

func (User) TableName() string {
	return "users"
}
