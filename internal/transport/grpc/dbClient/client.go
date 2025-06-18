package dbClient

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"homecloud-auth-service/internal/models"
	pb "homecloud-auth-service/internal/transport/grpc/protos"
)

type DBServiceClientImpl struct {
	client pb.DBServiceClient
	conn   *grpc.ClientConn
}

func NewDBServiceClient(host string, port int) (DBServiceClient, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	client := pb.NewDBServiceClient(conn)
	return &DBServiceClientImpl{
		client: client,
		conn:   conn,
	}, nil
}

func (c *DBServiceClientImpl) Connect() error {
	// Already connected in constructor
	return nil
}

func (c *DBServiceClientImpl) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *DBServiceClientImpl) CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	// Для тестирования возвращаем заглушку
	return uuid.New(), nil
}

func (c *DBServiceClientImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// Заглушка для тестирования
	return &models.User{
		ID:                 id,
		Email:             "test@example.com",
		Username:          "testuser",
		PasswordHash:      "$2a$12$hashedpassword",
		IsActive:          true,
		IsEmailVerified:   false,
		Role:              "user",
		StorageQuota:      10737418240, // 10GB
		UsedSpace:         0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		FailedLoginAttempts: 0,
		LockedUntil:       nil,
		LastLoginAt:       nil,
	}, nil
}

func (c *DBServiceClientImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Заглушка для тестирования
	return &models.User{
		ID:                 uuid.New(),
		Email:             email,
		Username:          "testuser",
		PasswordHash:      "$2a$12$hashedpassword",
		IsActive:          true,
		IsEmailVerified:   false,
		Role:              "user",
		StorageQuota:      10737418240, // 10GB
		UsedSpace:         0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		FailedLoginAttempts: 0,
		LockedUntil:       nil,
		LastLoginAt:       nil,
	}, nil
}

func (c *DBServiceClientImpl) UpdateUser(ctx context.Context, user *models.User) error {
	return nil // Заглушка успешного обновления
}

func (c *DBServiceClientImpl) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return nil // Заглушка успешного обновления
}

func (c *DBServiceClientImpl) UpdateUsername(ctx context.Context, id uuid.UUID, username string) error {
	return nil // Заглушка успешного обновления
}

func (c *DBServiceClientImpl) UpdateEmailVerification(ctx context.Context, id uuid.UUID, isVerified bool) error {
	return nil // Заглушка успешного обновления
}

func (c *DBServiceClientImpl) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return nil // Заглушка успешного обновления
}

func (c *DBServiceClientImpl) UpdateFailedLoginAttempts(ctx context.Context, id uuid.UUID, attempts int) error {
	return nil // Заглушка успешного обновления
}

func (c *DBServiceClientImpl) UpdateLockedUntil(ctx context.Context, id uuid.UUID, lockedUntil *time.Time) error {
	return nil // Заглушка успешного обновления
}

func (c *DBServiceClientImpl) UpdateStorageUsage(ctx context.Context, id uuid.UUID, usedSpace int64) error {
	return nil // Заглушка успешного обновления
}

func (c *DBServiceClientImpl) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	return false, nil // Заглушка - email не существует
}

func (c *DBServiceClientImpl) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	return false, nil // Заглушка - username не существует
}

// Вспомогательные функции для конвертации между protobuf и моделями
func userToProto(u *models.User) *pb.User {
	user := &pb.User{
		Id:                u.ID.String(),
		Email:            u.Email,
		Username:         u.Username,
		PasswordHash:     u.PasswordHash,
		IsActive:         u.IsActive,
		IsEmailVerified:  u.IsEmailVerified,
		Role:             u.Role,
		StorageQuota:     u.StorageQuota,
		UsedSpace:        u.UsedSpace,
		CreatedAt:        timestamppb.New(u.CreatedAt),
		UpdatedAt:        timestamppb.New(u.UpdatedAt),
		FailedLoginAttempts: int32(u.FailedLoginAttempts),
	}
	if u.LockedUntil != nil {
		user.LockedUntil = timestamppb.New(*u.LockedUntil)
	}
	if u.LastLoginAt != nil {
		user.LastLogin = timestamppb.New(*u.LastLoginAt)
	}
	return user
}

func protoToUser(p *pb.User) (*models.User, error) {
	id, err := uuid.Parse(p.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %v", err)
	}
	user := &models.User{
		ID:                 id,
		Email:             p.Email,
		Username:          p.Username,
		PasswordHash:      p.PasswordHash,
		IsActive:          p.IsActive,
		IsEmailVerified:   p.IsEmailVerified,
		Role:              p.Role,
		StorageQuota:      p.StorageQuota,
		UsedSpace:         p.UsedSpace,
		CreatedAt:         p.CreatedAt.AsTime(),
		UpdatedAt:         p.UpdatedAt.AsTime(),
		FailedLoginAttempts: int(p.FailedLoginAttempts),
	}
	if p.LockedUntil != nil {
		lockedUntil := p.LockedUntil.AsTime()
		user.LockedUntil = &lockedUntil
	}
	if p.LastLogin != nil {
		lastLogin := p.LastLogin.AsTime()
		user.LastLoginAt = &lastLogin
	}
	return user, nil
} 