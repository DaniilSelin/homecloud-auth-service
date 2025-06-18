package authServer

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"homecloud-auth-service/internal/interfaces"
	"homecloud-auth-service/internal/logger"
	"homecloud-auth-service/internal/models"
	"homecloud-auth-service/internal/security"
	pb "homecloud-auth-service/internal/transport/grpc/protos"
)

type AuthServiceServerImpl struct {
	pb.UnimplementedAuthServiceServer
	userService interfaces.UserService
	security    interfaces.Security
	logger      *logger.Logger
	server      *grpc.Server
	port        int
}

func NewAuthServiceServer(
	userService interfaces.UserService,
	security interfaces.Security,
	logger *logger.Logger,
	port int,
) AuthServiceServer {
	return &AuthServiceServerImpl{
		userService: userService,
		security:    security,
		logger:      logger,
		port:        port,
	}
}

func (s *AuthServiceServerImpl) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.server = grpc.NewServer()
	pb.RegisterAuthServiceServer(s.server, s)
	reflection.Register(s.server)

	s.logger.Info("Starting gRPC auth server", "port", s.port)
	return s.server.Serve(lis)
}

func (s *AuthServiceServerImpl) Stop() error {
	if s.server != nil {
		s.server.GracefulStop()
	}
	return nil
}

// gRPC method implementations
func (s *AuthServiceServerImpl) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := s.userService.Register(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		User: userToProto(user),
	}, nil
}

func (s *AuthServiceServerImpl) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, token, err := s.userService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		User:  userToProto(user),
		Token: token,
	}, nil
}

func (s *AuthServiceServerImpl) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	userID, err := parseUUID(req.UserId)
	if err != nil {
		return nil, err
	}

	user, err := s.userService.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserProfileResponse{
		User: userToProto(user),
	}, nil
}

func (s *AuthServiceServerImpl) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
	userID, err := parseUUID(req.UserId)
	if err != nil {
		return nil, err
	}

	var username, oldPassword, newPassword *string
	if req.Username != "" {
		username = &req.Username
	}
	if req.OldPassword != "" {
		oldPassword = &req.OldPassword
	}
	if req.NewPassword != "" {
		newPassword = &req.NewPassword
	}

	err = s.userService.UpdateUserProfile(ctx, userID, username, oldPassword, newPassword)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateUserProfileResponse{}, nil
}

func (s *AuthServiceServerImpl) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	err := s.userService.VerifyEmail(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	return &pb.VerifyEmailResponse{}, nil
}

func (s *AuthServiceServerImpl) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := s.userService.Logout(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	return &pb.LogoutResponse{}, nil
}

func (s *AuthServiceServerImpl) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	user, err := s.security.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	return &pb.ValidateTokenResponse{
		User: userToProto(user),
	}, nil
}

func (s *AuthServiceServerImpl) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	token, err := s.security.RefreshToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	return &pb.RefreshTokenResponse{
		Token: token,
	}, nil
}

// Helper functions
func userToProto(u *models.User) *pb.User {
	user := &pb.User{
		Id:                u.ID.String(),
		Email:            u.Email,
		Username:         u.Username,
		IsActive:         u.IsActive,
		IsEmailVerified:  u.IsEmailVerified,
		Role:             u.Role,
		StorageQuota:     u.StorageQuota,
		UsedSpace:        u.UsedSpace,
		FailedLoginAttempts: int32(u.FailedLoginAttempts),
	}

	// Note: We don't include password hash in the response
	return user
}

func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
} 