package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	authBiz "github.com/hobbyGG/Dawnix/internal/auth/biz"
	"github.com/hobbyGG/Dawnix/util"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Config struct {
	JWTSecret        string
	JWTIssuer        string
	JWTExpireMinutes int
}

type Service struct {
	repo   authBiz.Repo
	cfg    Config
	logger *zap.Logger
}

var ErrUsernameAlreadyExists = errors.New("username already exists")

type RegisterParams struct {
	Username    string
	Password    string
	DisplayName string
}

type RegisterResult struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type LoginParams struct {
	Username string
	Password string
}

type LoginResult struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type Claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func NewService(repo authBiz.Repo, cfg Config, logger *zap.Logger) *Service {
	if cfg.JWTExpireMinutes <= 0 {
		cfg.JWTExpireMinutes = 120
	}
	return &Service{repo: repo, cfg: cfg, logger: logger}
}

func (s *Service) Register(ctx context.Context, params *RegisterParams) (*RegisterResult, error) {
	if params == nil {
		return nil, fmt.Errorf("register params is nil")
	}
	username := params.Username
	if username == "" || params.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}
	displayName := params.DisplayName
	if displayName == "" {
		displayName = username
	}

	// 检查该 provider 下是否已经有相同用户名的账号存在
	_, err := s.repo.GetIdentityByProviderAndSub(ctx, authBiz.ProviderLocalPassword, username)
	if err == nil {
		return nil, ErrUsernameAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check username existence failed: %w", err)
	}

	// 对密码进行加密
	credentialHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}
	userID := util.NextSnowflakeIDString()
	if err := s.repo.CreateUserAndIdentity(ctx, &authBiz.User{
		UserID:      userID,
		DisplayName: displayName,
		Status:      authBiz.UserStatusActive,
	}, &authBiz.AuthIdentity{
		UserID:         userID,
		Provider:       authBiz.ProviderLocalPassword,
		ProviderSub:    username,
		CredentialHash: string(credentialHash),
	}); err != nil {
		return nil, fmt.Errorf("register user failed: %w", err)
	}

	return &RegisterResult{
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
	}, nil
}

func (s *Service) Login(ctx context.Context, params *LoginParams) (*LoginResult, error) {
	if params == nil {
		return nil, fmt.Errorf("login params is nil")
	}
	if params.Username == "" || params.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	identity, err := s.repo.GetIdentityByProviderAndSub(ctx, authBiz.ProviderLocalPassword, params.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invalid username or password")
		}
		return nil, fmt.Errorf("get auth identity failed: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(identity.CredentialHash), []byte(params.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	user, err := s.repo.GetUserByID(ctx, identity.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user failed: %w", err)
	}
	if user.Status != authBiz.UserStatusActive {
		return nil, fmt.Errorf("user is not active")
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(s.cfg.JWTExpireMinutes) * time.Minute)
	claims := &Claims{
		UserID: user.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.JWTIssuer,
			Subject:   user.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("sign token failed: %w", err)
	}

	if err := s.repo.UpdateIdentityLastLogin(ctx, identity.ID, now.Unix()); err != nil {
		s.logger.Warn("update identity last login failed", zap.Error(err))
	}

	return &LoginResult{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *Service) Logout(ctx context.Context) error {
	_ = ctx
	return nil
}

func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token is required")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token failed: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("token subject is empty")
	}
	return claims, nil
}
