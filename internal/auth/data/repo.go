package data

import (
	"context"
	"fmt"
	"time"

	authBiz "github.com/hobbyGG/Dawnix/internal/auth/biz"
	authModel "github.com/hobbyGG/Dawnix/internal/auth/data/model"
	coreData "github.com/hobbyGG/Dawnix/internal/workflow/data"
	"gorm.io/gorm"
)

type Repo struct {
	db *coreData.Data
}

func NewRepo(db *coreData.Data) authBiz.Repo {
	return &Repo{db: db}
}

func (r *Repo) GetUserByID(ctx context.Context, userID string) (*authBiz.User, error) {
	var user authModel.User
	if err := r.db.DB(ctx).WithContext(ctx).First(&user, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("get user by id failed: %w", err)
	}
	return userToDomain(&user), nil
}

func (r *Repo) SearchActiveUsers(ctx context.Context, keyword string, limit int) ([]*authBiz.User, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.DB(ctx).WithContext(ctx).
		Model(&authModel.User{}).
		Where("status = ?", authBiz.UserStatusActive)
	if keyword != "" {
		likeValue := "%" + keyword + "%"
		query = query.Where("display_name ILIKE ? OR user_id ILIKE ?", likeValue, likeValue)
	}

	var users []authModel.User
	if err := query.Order("display_name ASC, user_id ASC").Limit(limit).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("search active users failed: %w", err)
	}

	result := make([]*authBiz.User, 0, len(users))
	for i := range users {
		result = append(result, userToDomain(&users[i]))
	}
	return result, nil
}

func (r *Repo) GetIdentityByProviderAndSub(ctx context.Context, provider string, providerSub string) (*authBiz.AuthIdentity, error) {
	var identity authModel.AuthIdentity
	err := r.db.DB(ctx).WithContext(ctx).
		Where("provider = ? AND provider_sub = ?", provider, providerSub).
		First(&identity).Error
	if err != nil {
		return nil, fmt.Errorf("get identity by provider and subject failed: %w", err)
	}
	return identityToDomain(&identity), nil
}

func (r *Repo) CreateUserAndIdentity(ctx context.Context, user *authBiz.User, identity *authBiz.AuthIdentity) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	if identity == nil {
		return fmt.Errorf("identity is nil")
	}

	userPO := userToPO(user)
	identityPO := identityToPO(identity)

	if err := r.db.DB(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(userPO).Error; err != nil {
			return fmt.Errorf("create user failed: %w", err)
		}
		if err := tx.Create(identityPO).Error; err != nil {
			return fmt.Errorf("create identity failed: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("create user and identity failed: %w", err)
	}
	return nil
}

func userToDomain(user *authModel.User) *authBiz.User {
	if user == nil {
		return nil
	}
	return &authBiz.User{
		UserID:      user.UserID,
		DisplayName: user.DisplayName,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func identityToDomain(identity *authModel.AuthIdentity) *authBiz.AuthIdentity {
	if identity == nil {
		return nil
	}
	return &authBiz.AuthIdentity{
		ID:             identity.ID,
		UserID:         identity.UserID,
		Provider:       identity.Provider,
		ProviderSub:    identity.ProviderSub,
		CredentialHash: identity.CredentialHash,
		LastLoginAt:    identity.LastLoginAt,
	}
}

func userToPO(user *authBiz.User) *authModel.User {
	if user == nil {
		return nil
	}
	return &authModel.User{
		UserID:      user.UserID,
		DisplayName: user.DisplayName,
		Status:      user.Status,
	}
}

func identityToPO(identity *authBiz.AuthIdentity) *authModel.AuthIdentity {
	if identity == nil {
		return nil
	}
	return &authModel.AuthIdentity{
		UserID:         identity.UserID,
		Provider:       identity.Provider,
		ProviderSub:    identity.ProviderSub,
		CredentialHash: identity.CredentialHash,
	}
}

func (r *Repo) UpdateIdentityLastLogin(ctx context.Context, id int64, loginAtUnix int64) error {
	loginAt := time.Unix(loginAtUnix, 0)
	result := r.db.DB(ctx).WithContext(ctx).Model(&authModel.AuthIdentity{}).Where("id = ?", id).Update("last_login_at", loginAt)
	if result.Error != nil {
		return fmt.Errorf("update identity last login failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
