package interfaces

import (
	"context"
	"github.com/google/uuid"
	"homecloud-auth-service/internal/models"
)

type UserService interface {
	// Аутентификация
	Register(ctx context.Context, email, username, password string) (*models.User, string, error)
	Login(ctx context.Context, email, password string) (*models.User, string, error)
	ValidateToken(ctx context.Context, token string) (*models.User, error)
	Logout(ctx context.Context, token string) error
	
	// Профиль пользователя
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, username *string, oldPassword *string, newPassword *string) error
	
	// Верификация email
	VerifyEmail(ctx context.Context, token string) error
	SendVerificationEmail(ctx context.Context, userID uuid.UUID) error
	
	// Управление аккаунтом
	UpdateStorageUsage(ctx context.Context, userID uuid.UUID, usedSpace int64) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type AuthService interface {
	GenerateToken(userID uuid.UUID) (string, error)
	ValidateToken(token string) (*TokenClaims, error)
	RefreshToken(token string) (string, error)
	InvalidateToken(token string) error
}

type TokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	TokenID string   `json:"token_id,omitempty"`
} 