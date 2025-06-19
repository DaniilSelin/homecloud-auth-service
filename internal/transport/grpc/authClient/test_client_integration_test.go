package authClient

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthServiceIntegration(t *testing.T) {
	client, err := NewAuthTestClient("127.0.0.1", 50051)
	assert.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Тест регистрации
	email := "integration_test@example.com"
	username := "integration_user"
	password := "integration_password"
	user, err := client.Register(ctx, email, username, password)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, username, user.Username)

	// Тест логина
	loggedUser, token, err := client.Login(ctx, email, password)
	assert.NoError(t, err)
	assert.NotNil(t, loggedUser)
	assert.NotEmpty(t, token)
	assert.Equal(t, email, loggedUser.Email)

	// Тест получения профиля
	profile, err := client.GetUserProfile(ctx, loggedUser.Id)
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, email, profile.Email)
}
