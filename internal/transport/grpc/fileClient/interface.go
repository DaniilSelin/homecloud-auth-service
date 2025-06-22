package fileClient

import (
	"context"
)

// FileServiceClient интерфейс для работы с файловым сервисом
type FileServiceClient interface {
	// CreateUserDirectory создает домашнюю директорию для пользователя
	CreateUserDirectory(ctx context.Context, userID, username string) (bool, string, string, error)

	// Close закрывает соединение с сервисом
	Close() error
}
