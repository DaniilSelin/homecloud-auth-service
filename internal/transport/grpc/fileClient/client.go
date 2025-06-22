package fileClient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"homecloud-auth-service/internal/transport/grpc/fileClient/protos"
)

// FileServiceClientImpl реализация клиента файлового сервиса
type FileServiceClientImpl struct {
	conn   *grpc.ClientConn
	client protos.FileServiceClient
}

// NewFileServiceClient создает новый клиент файлового сервиса
func NewFileServiceClient(host string, port int) (*FileServiceClientImpl, error) {
	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to file service: %w", err)
	}

	client := protos.NewFileServiceClient(conn)

	return &FileServiceClientImpl{
		conn:   conn,
		client: client,
	}, nil
}

// CreateUserDirectory создает домашнюю директорию для пользователя
func (c *FileServiceClientImpl) CreateUserDirectory(ctx context.Context, userID, username string) (bool, string, string, error) {
	// Добавляем таймаут к контексту
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := &protos.CreateUserDirectoryRequest{
		UserId:   userID,
		Username: username,
	}

	response, err := c.client.CreateUserDirectory(ctx, request)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to create user directory: %w", err)
	}

	return response.Success, response.Message, response.DirectoryPath, nil
}

// Close закрывает соединение с сервисом
func (c *FileServiceClientImpl) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected проверяет состояние соединения
func (c *FileServiceClientImpl) IsConnected() bool {
	if c.conn == nil {
		return false
	}
	return c.conn.GetState() == connectivity.Ready
}
