package service

import (
	"context"
	"errors"
	"time"

	"huihua/finance/internal/config"
	"huihua/finance/internal/repository"
	"huihua/finance/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg}
}

type LoginResult struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}
	if !user.IsActive {
		return nil, errors.New("user is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	expiry, _ := time.ParseDuration(s.cfg.JWT.Expiry)
	token, err := jwt.GenerateToken(s.cfg.JWT.Secret, user.ID, user.TenantID, user.Role, expiry)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:     token,
		UserID:    user.ID.String(),
		TenantID:  user.TenantID.String(),
		Role:      user.Role,
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}
