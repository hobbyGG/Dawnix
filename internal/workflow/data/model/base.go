package model

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
)

type BaseModel struct {
	ID        int64                  `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt gorm.DeletedAt         `gorm:"index" json:"-"`
	CreatedBy string                 `gorm:"type:varchar(64)" json:"created_by"`
	UpdatedBy string                 `gorm:"type:varchar(64)" json:"updated_by"`
	Revision  optimisticlock.Version `json:"revision"`
}
