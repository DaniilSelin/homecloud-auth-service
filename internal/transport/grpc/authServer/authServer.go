package authServer

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"homecloud-auth-service/config"
	"homecloud-auth-service/internal/interfaces"
	"homecloud-auth-service/internal/security"
	pb "homecloud-auth-service/internal/transport/grpc/protos"

	"github.com/google/uuid"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	userService interfaces.UserService
	sec         *security.Security
	cfg         *config.GrpcConfig
	contxt      *context.Context
}

func NewAuthServer(ctx *context.Context, userService interfaces.UserService, sec *security.Security, cfg *config.GrpcConfig) *AuthServer {
	return &AuthServer{
		userService: userService,
		sec:         sec,
		cfg:         cfg,
		contxt:      ctx,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, _, err := s.userService.Register(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("registration failed: %v", err)
	}

	return &pb.RegisterResponse{
		User: &pb.AuthUser{
			Id:              user.ID.String(),
			Email:           user.Email,
			Username:        user.Username,
			IsActive:        user.IsActive,
			IsEmailVerified: user.IsEmailVerified,
			StorageQuota:    user.StorageQuota,
			UsedSpace:       user.UsedSpace,
			Role:            user.Role,
			IsAdmin:         user.IsAdmin,
			CreatedAt:       user.CreatedAt.String(),
			UpdatedAt:       user.UpdatedAt.String(),
		},
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, token, err := s.userService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}

	return &pb.LoginResponse{
		User: &pb.AuthUser{
			Id:              user.ID.String(),
			Email:           user.Email,
			Username:        user.Username,
			IsActive:        user.IsActive,
			IsEmailVerified: user.IsEmailVerified,
			StorageQuota:    user.StorageQuota,
			UsedSpace:       user.UsedSpace,
			Role:            user.Role,
			IsAdmin:         user.IsAdmin,
			CreatedAt:       user.CreatedAt.String(),
			UpdatedAt:       user.UpdatedAt.String(),
		},
		Token: token,
	}, nil
}

func (s *AuthServer) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	user, err := s.userService.GetUserProfile(ctx, parseUUID(req.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %v", err)
	}

	return &pb.GetUserProfileResponse{
		User: &pb.AuthUser{
			Id:              user.ID.String(),
			Email:           user.Email,
			Username:        user.Username,
			IsActive:        user.IsActive,
			IsEmailVerified: user.IsEmailVerified,
			StorageQuota:    user.StorageQuota,
			UsedSpace:       user.UsedSpace,
			Role:            user.Role,
			IsAdmin:         user.IsAdmin,
			CreatedAt:       user.CreatedAt.String(),
			UpdatedAt:       user.UpdatedAt.String(),
		},
	}, nil
}

func (s *AuthServer) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
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

	err := s.userService.UpdateProfile(ctx, parseUUID(req.UserId), username, oldPassword, newPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %v", err)
	}

	return &pb.UpdateUserProfileResponse{}, nil
}

func (s *AuthServer) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	err := s.userService.VerifyEmail(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("email verification failed: %v", err)
	}

	return &pb.VerifyEmailResponse{}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := s.userService.Logout(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("logout failed: %v", err)
	}

	return &pb.LogoutResponse{}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := s.sec.ValidateToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %v", err)
	}

	return &pb.ValidateTokenResponse{
		User: &pb.AuthUser{
			Id: claims.UserID.String(),
		},
	}, nil
}

func parseUUID(id string) uuid.UUID {
	u, _ := uuid.Parse(id)
	return u
}

func (a *AuthServer) StartAuthServer() error {
	port := fmt.Sprintf(":%d", a.cfg.Port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, a)

	reflection.Register(grpcServer)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		fmt.Printf("gRPC auth server started on %s\n", port)
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Printf("failed to serve: %v\n", err)
			panic(err)
		}
	}()

	<-stop
	fmt.Printf("gRPC auth server is shutting down...\n")
	grpcServer.GracefulStop()
	fmt.Printf("gRPC auth server stopped on %s\n", port)
	return nil
}
