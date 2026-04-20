package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	authBiz "github.com/hobbyGG/Dawnix/internal/auth/biz"
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
