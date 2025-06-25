package fileClient

import (
	"context"
	"fmt"
)

// MockFileServiceClient - заглушка для тестирования
type MockFileServiceClient struct {
	shouldFail    bool
	directoryPath string
}

// NewMockFileServiceClient создает новый mock клиент
func NewMockFileServiceClient(shouldFail bool) *MockFileServiceClient {
	return &MockFileServiceClient{
		shouldFail:    shouldFail,
		directoryPath: "/home/users",
	}
}

// CreateUserDirectory создает домашнюю директорию для пользователя (mock)
func (m *MockFileServiceClient) CreateUserDirectory(ctx context.Context, userID, username string) (bool, string, string, error) {
	if m.shouldFail {
		return false, "Mock: Directory creation failed", "", fmt.Errorf("mock directory creation error")
	}

	// Симулируем успешное создание директории
	path := fmt.Sprintf("%s/%s", m.directoryPath, userID)
	return true, "Mock: Directory created successfully", path, nil
}

// Close закрывает соединение с сервисом (mock)
func (m *MockFileServiceClient) Close() error {
	return nil
}

// IsConnected проверяет состояние соединения (mock)
func (m *MockFileServiceClient) IsConnected() bool {
	return !m.shouldFail
}
