package authClient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "homecloud-auth-service/internal/transport/grpc/protos"
)

type AuthTestClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthTestClient(host string, port int) (*AuthTestClient, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	client := pb.NewAuthServiceClient(conn)
	return &AuthTestClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *AuthTestClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *AuthTestClient) Register(ctx context.Context, email, username, password string) (*pb.AuthUser, error) {
	req := &pb.RegisterRequest{
		Email:    email,
		Username: username,
		Password: password,
	}

	resp, err := c.client.Register(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("register failed: %v", err)
	}

	fmt.Printf("User registered successfully: %v\n", resp.User)
	return resp.User, nil
}

func (c *AuthTestClient) Login(ctx context.Context, email, password string) (*pb.AuthUser, string, error) {
	req := &pb.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.client.Login(ctx, req)
	if err != nil {
		return nil, "", fmt.Errorf("login failed: %v", err)
	}

	fmt.Printf("User logged in successfully: %v\n", resp.User)
	return resp.User, resp.Token, nil
}

func (c *AuthTestClient) GetUserProfile(ctx context.Context, userID string) (*pb.AuthUser, error) {
	req := &pb.GetUserProfileRequest{
		UserId: userID,
	}

	resp, err := c.client.GetUserProfile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get user profile failed: %v", err)
	}

	fmt.Printf("User profile: %v\n", resp.User)
	return resp.User, nil
}

func (c *AuthTestClient) ValidateToken(ctx context.Context, token string) (*pb.AuthUser, error) {
	req := &pb.ValidateTokenRequest{
		Token: token,
	}

	resp, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %v", err)
	}

	fmt.Printf("Token valid, user: %v\n", resp.User)
	return resp.User, nil
}
