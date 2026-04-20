package biz

import "time"

const (
	ProviderLocalPassword = "local_password"

	UserStatusActive   = "ACTIVE"
	UserStatusDisabled = "DISABLED"
)

type User struct {
	UserID      string
	DisplayName string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AuthIdentity struct {
	ID             int64
	UserID         string
	Provider       string
	ProviderSub    string
	CredentialHash string
	LastLoginAt    *time.Time
}

type Principal struct {
	UserID string
}
