package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuthIdentity struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Provider       string         `gorm:"type:varchar(64);not null;uniqueIndex:uk_provider_sub,priority:1" json:"provider"`
	ProviderSub    string         `gorm:"type:varchar(128);not null;uniqueIndex:uk_provider_sub,priority:2" json:"provider_subject"`
	CredentialHash string         `gorm:"type:varchar(255)" json:"-"`
	Meta           datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"meta"`
	LastLoginAt    *time.Time     `json:"last_login_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AuthIdentity) TableName() string {
	return "auth_identities"
}
