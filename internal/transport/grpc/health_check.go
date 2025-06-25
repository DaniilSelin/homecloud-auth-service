package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// checkServiceHealth проверяет доступность gRPC сервиса
func checkServiceHealth(host string, port int, serviceName string) error {
	address := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Attempting to connect to %s at %s...\n", serviceName, address)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect to %s at %s: %v\n", serviceName, address, err)
		return fmt.Errorf("failed to connect to %s: %w", serviceName, err)
	}
	defer conn.Close()

	// Проверяем состояние соединения
	state := conn.GetState()
	if state != connectivity.Ready {
		fmt.Printf("Failed to connect to %s at %s: %v\n", serviceName, address, err)
		return fmt.Errorf("%s is not ready (state: %s)", serviceName, state)
	}

	fmt.Printf("Successfully connected to %s at %s (State: %s)\n", serviceName, address, state)
	return nil
}

// CheckAllServices проверяет доступность всех gRPC сервисов
func CheckAllServices(dbHost string, dbPort int, fileHost string, filePort int) error {
	fmt.Println("Starting gRPC services health check...")
	fmt.Println("==================================================")

	// Проверяем DB Manager
	fmt.Printf("Checking DB Manager service...\n")
	if err := checkServiceHealth(dbHost, dbPort, "DB Manager"); err != nil {
		fmt.Printf("DB Manager health check failed: %v\n", err)
		return err
	}

	// Проверяем File Service
	fmt.Printf("Checking File Service...\n")
	if err := checkServiceHealth(fileHost, filePort, "File Service"); err != nil {
		fmt.Printf("File Service health check failed: %v\n", err)
		return err
	}

	fmt.Println("==================================================")
	fmt.Println("All gRPC services are available and ready!")
	fmt.Println("==================================================")

	return nil
}
