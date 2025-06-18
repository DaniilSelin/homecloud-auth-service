package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"homecloud-auth-service/internal/transport/grpc/authClient"
)

func main() {
	// Создаем тестового клиента
	client, err := authClient.NewAuthTestClient("localhost", 50051)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx := context.Background()

	// Тестируем регистрацию
	registeredUser, err := client.Register(ctx, "test@example.com", "testuser", "password123")
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Registered user ID: %s\n", registeredUser.GetId())

	// Делаем небольшую паузу
	time.Sleep(time.Second)

	// Тестируем логин
	loggedInUser, token, err := client.Login(ctx, "test@example.com", "password123")
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Logged in user ID: %s\n", loggedInUser.GetId())
	fmt.Printf("Auth token: %s\n", token)

	// Тестируем получение профиля
	profile, err := client.GetUserProfile(ctx, loggedInUser.GetId())
	if err != nil {
		fmt.Printf("Get profile failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Retrieved profile: %v\n", profile)

	// Тестируем валидацию токена
	validatedUser, err := client.ValidateToken(ctx, token)
	if err != nil {
		fmt.Printf("Token validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Validated user ID: %s\n", validatedUser.GetId())

	fmt.Println("All tests passed successfully!")
} 