package biz

import "context"

type Repo interface {
	GetUserByID(ctx context.Context, userID string) (*User, error)
	SearchActiveUsers(ctx context.Context, keyword string, limit int) ([]*User, error)
	GetIdentityByProviderAndSub(ctx context.Context, provider string, providerSub string) (*AuthIdentity, error)
	CreateUserAndIdentity(ctx context.Context, user *User, identity *AuthIdentity) error
	UpdateIdentityLastLogin(ctx context.Context, id int64, loginAtUnix int64) error
}
