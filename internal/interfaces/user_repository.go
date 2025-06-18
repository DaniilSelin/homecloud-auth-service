package interfaces

import (
	"context"
	"time"
	"github.com/google/uuid"
	"homecloud-auth-service/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateUsername(ctx context.Context, id uuid.UUID, username string) error
	UpdateEmailVerification(ctx context.Context, id uuid.UUID, isVerified bool) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateFailedLoginAttempts(ctx context.Context, id uuid.UUID, attempts int) error
	UpdateLockedUntil(ctx context.Context, id uuid.UUID, lockedUntil *time.Time) error
	UpdateStorageUsage(ctx context.Context, id uuid.UUID, usedSpace int64) error
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
} 