package authServer

import (
	"context"

	"github.com/google/uuid"
	"homecloud-auth-service/internal/models"
)

// AuthServiceServer определяет интерфейс для gRPC сервера авторизации
type AuthServiceServer interface {
	// User operations
	Register(ctx context.Context, email, username, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.User, string, error) // user, token, error
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, username, oldPassword, newPassword *string) error
	VerifyEmail(ctx context.Context, token string) error
	Logout(ctx context.Context, token string) error

	// Token operations
	ValidateToken(ctx context.Context, token string) (*models.User, error)
	RefreshToken(ctx context.Context, token string) (string, error)

	// Server management
	Start() error
	Stop() error
} 