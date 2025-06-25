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

// NewDBServiceClient создает новый клиент для взаимодействия с сервисом БД
func NewDBServiceClient(host string, port int) (*DBServiceClientImpl, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Connecting to DB Manager at %s...\n", addr)

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect to DB Manager at %s: %v\n", addr, err)
		return nil, fmt.Errorf("failed to connect to db manager: %w", err)
	}

	client := pb.NewDBServiceClient(conn)
	fmt.Printf("Successfully connected to DB Manager at %s\n", addr)

	return &DBServiceClientImpl{
		conn:   conn,
		client: client,
	}, nil
}

func (c *DBServiceClientImpl) Connect() error {
	return nil
}

func (c *DBServiceClientImpl) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *DBServiceClientImpl) CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	req := &pb.User{
		Id:                  user.ID.String(),
		Email:               user.Email,
		Username:            user.Username,
		PasswordHash:        user.PasswordHash,
		IsActive:            user.IsActive,
		IsEmailVerified:     user.IsEmailVerified,
		Role:                user.Role,
		StorageQuota:        user.StorageQuota,
		UsedSpace:           user.UsedSpace,
		CreatedAt:           timestamppb.New(user.CreatedAt),
		UpdatedAt:           timestamppb.New(user.UpdatedAt),
		FailedLoginAttempts: int32(user.FailedLoginAttempts),
	}
	if user.LockedUntil != nil {
		req.LockedUntil = timestamppb.New(*user.LockedUntil)
	}
	if user.LastLoginAt != nil {
		req.LastLogin = timestamppb.New(*user.LastLoginAt)
	}
	resp, err := c.client.CreateUser(ctx, req)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(resp.Id)
}

func (c *DBServiceClientImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	req := &pb.UserID{Id: id.String()}
	resp, err := c.client.GetUserByID(ctx, req)
	if err != nil {
		return nil, err
	}
	return protoToUser(resp)
}

func (c *DBServiceClientImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	req := &pb.EmailRequest{Email: email}
	resp, err := c.client.GetUserByEmail(ctx, req)
	if err != nil {
		return nil, err
	}
	return protoToUser(resp)
}

func (c *DBServiceClientImpl) UpdateUser(ctx context.Context, user *models.User) error {
	req := &pb.User{
		Id:                  user.ID.String(),
		Email:               user.Email,
		Username:            user.Username,
		PasswordHash:        user.PasswordHash,
		IsActive:            user.IsActive,
		IsEmailVerified:     user.IsEmailVerified,
		Role:                user.Role,
		StorageQuota:        user.StorageQuota,
		UsedSpace:           user.UsedSpace,
		CreatedAt:           timestamppb.New(user.CreatedAt),
		UpdatedAt:           timestamppb.New(user.UpdatedAt),
		FailedLoginAttempts: int32(user.FailedLoginAttempts),
	}
	if user.LockedUntil != nil {
		req.LockedUntil = timestamppb.New(*user.LockedUntil)
	}
	if user.LastLoginAt != nil {
		req.LastLogin = timestamppb.New(*user.LastLoginAt)
	}
	_, err := c.client.UpdateUser(ctx, req)
	return err
}

func (c *DBServiceClientImpl) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	req := &pb.UpdatePasswordRequest{Id: id.String(), PasswordHash: passwordHash}
	_, err := c.client.UpdatePassword(ctx, req)
	return err
}

func (c *DBServiceClientImpl) UpdateUsername(ctx context.Context, id uuid.UUID, username string) error {
	req := &pb.UpdateUsernameRequest{Id: id.String(), Username: username}
	_, err := c.client.UpdateUsername(ctx, req)
	return err
}

func (c *DBServiceClientImpl) UpdateEmailVerification(ctx context.Context, id uuid.UUID, isVerified bool) error {
	req := &pb.UpdateEmailVerificationRequest{Id: id.String(), IsVerified: isVerified}
	_, err := c.client.UpdateEmailVerification(ctx, req)
	return err
}

func (c *DBServiceClientImpl) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	req := &pb.UserID{Id: id.String()}
	_, err := c.client.UpdateLastLogin(ctx, req)
	return err
}

func (c *DBServiceClientImpl) UpdateFailedLoginAttempts(ctx context.Context, id uuid.UUID, attempts int) error {
	req := &pb.UpdateFailedLoginAttemptsRequest{Id: id.String(), Attempts: int32(attempts)}
	_, err := c.client.UpdateFailedLoginAttempts(ctx, req)
	return err
}

func (c *DBServiceClientImpl) UpdateLockedUntil(ctx context.Context, id uuid.UUID, lockedUntil *time.Time) error {
	var ts *timestamppb.Timestamp
	if lockedUntil != nil {
		ts = timestamppb.New(*lockedUntil)
	}
	req := &pb.UpdateLockedUntilRequest{Id: id.String(), LockedUntil: ts}
	_, err := c.client.UpdateLockedUntil(ctx, req)
	return err
}

func (c *DBServiceClientImpl) UpdateStorageUsage(ctx context.Context, id uuid.UUID, usedSpace int64) error {
	req := &pb.UpdateStorageUsageRequest{Id: id.String(), UsedSpace: usedSpace}
	_, err := c.client.UpdateStorageUsage(ctx, req)
	return err
}

func (c *DBServiceClientImpl) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	req := &pb.EmailRequest{Email: email}
	resp, err := c.client.CheckEmailExists(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.Exists, nil
}

func (c *DBServiceClientImpl) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	req := &pb.UsernameRequest{Username: username}
	resp, err := c.client.CheckUsernameExists(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.Exists, nil
}

func protoToUser(p *pb.User) (*models.User, error) {
	id, err := uuid.Parse(p.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %v", err)
	}
	user := &models.User{
		ID:                  id,
		Email:               p.Email,
		Username:            p.Username,
		PasswordHash:        p.PasswordHash,
		IsActive:            p.IsActive,
		IsEmailVerified:     p.IsEmailVerified,
		Role:                p.Role,
		StorageQuota:        p.StorageQuota,
		UsedSpace:           p.UsedSpace,
		CreatedAt:           p.CreatedAt.AsTime(),
		UpdatedAt:           p.UpdatedAt.AsTime(),
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
