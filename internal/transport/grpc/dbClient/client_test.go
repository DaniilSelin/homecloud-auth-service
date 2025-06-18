package dbClient

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"homecloud-auth-service/internal/models"
)

func TestDBServiceClient(t *testing.T) {
	// Создаем клиент с заглушками
	client, err := NewDBServiceClient("localhost", 50051)
	assert.NoError(t, err)
	defer client.Close()

	// Тестируем создание пользователя
	user := &models.User{
		Email:       "test@example.com",
		Username:    "testuser",
		PasswordHash: "$2a$12$hashedpassword",
	}

	ctx := context.Background()
	userID, err := client.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.NotEmpty(t, userID)

	// Тестируем получение пользователя по ID
	fetchedUser, err := client.GetUserByID(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedUser)
	assert.Equal(t, "test@example.com", fetchedUser.Email)

	// Тестируем получение пользователя по email
	emailUser, err := client.GetUserByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, emailUser)
	assert.Equal(t, "testuser", emailUser.Username)

	// Тестируем обновление пользователя
	emailUser.Username = "newusername"
	err = client.UpdateUser(ctx, emailUser)
	assert.NoError(t, err)

	// Тестируем проверку существования email
	exists, err := client.CheckEmailExists(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.False(t, exists) // Поскольку это заглушка, она всегда возвращает false

	// Тестируем обновление пароля
	err = client.UpdatePassword(ctx, userID, "newhash")
	assert.NoError(t, err)

	// Тестируем верификацию email
	err = client.UpdateEmailVerification(ctx, userID, true)
	assert.NoError(t, err)

	// Тестируем обновление попыток входа
	err = client.UpdateFailedLoginAttempts(ctx, userID, 1)
	assert.NoError(t, err)

	// Тестируем блокировку
	lockTime := time.Now().Add(time.Hour)
	err = client.UpdateLockedUntil(ctx, userID, &lockTime)
	assert.NoError(t, err)

	// Тестируем обновление использования хранилища
	err = client.UpdateStorageUsage(ctx, userID, 1024)
	assert.NoError(t, err)
} 