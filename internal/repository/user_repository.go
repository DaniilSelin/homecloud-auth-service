package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"homecloud-auth-service/internal/models"
	"homecloud-auth-service/internal/transport/grpc/dbClient"
)

// UserRepository реализует доступ к данным пользователей через gRPC
type UserRepository struct {
	dbClient dbClient.DBServiceClient
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(dbClient dbClient.DBServiceClient) *UserRepository {
	return &UserRepository{
		dbClient: dbClient,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	return r.dbClient.CreateUser(ctx, user)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return r.dbClient.GetUserByID(ctx, id)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.dbClient.GetUserByEmail(ctx, email)
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.dbClient.UpdateUser(ctx, user)
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.dbClient.UpdatePassword(ctx, id, passwordHash)
}

func (r *UserRepository) UpdateUsername(ctx context.Context, id uuid.UUID, username string) error {
	return r.dbClient.UpdateUsername(ctx, id, username)
}

func (r *UserRepository) UpdateEmailVerification(ctx context.Context, id uuid.UUID, isVerified bool) error {
	return r.dbClient.UpdateEmailVerification(ctx, id, isVerified)
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.dbClient.UpdateLastLogin(ctx, id)
}

func (r *UserRepository) UpdateFailedLoginAttempts(ctx context.Context, id uuid.UUID, attempts int) error {
	return r.dbClient.UpdateFailedLoginAttempts(ctx, id, attempts)
}

func (r *UserRepository) UpdateLockedUntil(ctx context.Context, id uuid.UUID, lockedUntil *time.Time) error {
	return r.dbClient.UpdateLockedUntil(ctx, id, lockedUntil)
}

func (r *UserRepository) UpdateStorageUsage(ctx context.Context, id uuid.UUID, usedSpace int64) error {
	return r.dbClient.UpdateStorageUsage(ctx, id, usedSpace)
}

func (r *UserRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	return r.dbClient.CheckEmailExists(ctx, email)
}

func (r *UserRepository) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	return r.dbClient.CheckUsernameExists(ctx, username)
}

// Вспомогательная функция для конвертации protobuf пользователя в интерфейс
// func convertPBUserToUser(pbUser *pb.User) *interfaces.User {
//     return &interfaces.User{
//         ID:              uuid.MustParse(pbUser.Id),
//         Email:           pbUser.Email,
//         Username:        pbUser.Username,
//         PasswordHash:    pbUser.PasswordHash,
//         IsActive:        pbUser.IsActive,
//         IsEmailVerified: pbUser.IsEmailVerified,
//         Role:            pbUser.Role,
//         StorageQuota:    pbUser.StorageQuota,
//         UsedSpace:       pbUser.UsedSpace,
//         CreatedAt:       pbUser.CreatedAt.AsTime(),
//         UpdatedAt:       pbUser.UpdatedAt.AsTime(),
//     }
// }