package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"homecloud-auth-service/config"
	"homecloud-auth-service/internal/logger"
	"homecloud-auth-service/internal/repository"
	"homecloud-auth-service/internal/security"
	"homecloud-auth-service/internal/service"
	"homecloud-auth-service/internal/transport/grpc/authServer"
	"homecloud-auth-service/internal/transport/grpc/dbClient"
	"homecloud-auth-service/internal/transport/grpc/fileClient"
	"homecloud-auth-service/internal/transport/http/api"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	srv, logBase, err := run(ctx, os.Stdout, os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	<-ctx.Done()
	logBase.Info(ctx, "Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Gracefully stop HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logBase.Error(ctx, "HTTP server shutdown failed", zap.Error(err))
	}
	logBase.Info(ctx, "Servers exited gracefully")
}

func run(ctx context.Context, w io.Writer, args []string) (*http.Server, *logger.Logger, error) {
	// Конфиг и логгер
	cfg, err := config.LoadConfig("config/config.local.yaml")
	if err != nil {
		return nil, nil, err
	}
	logBase, err := logger.New(cfg)
	if err != nil {
		return nil, nil, err
	}
	ctx = logger.CtxWWithLogger(ctx, logBase)

	// Создаём gRPC dbClient
	fmt.Printf("Creating DB Manager client connection to %s:%d...\n", cfg.DbManager.Host, cfg.DbManager.Port)
	dbClient, err := dbClient.NewDBServiceClient(cfg.DbManager.Host, cfg.DbManager.Port)
	if err != nil {
		fmt.Printf("Failed to create DB Manager client: %v\n", err)
		return nil, nil, fmt.Errorf("failed to create dbClient: %w", err)
	}
	fmt.Printf("DB Manager client created successfully\n")

	// Создаём репозиторий
	userRepo := repository.NewUserRepository(dbClient)
	fmt.Printf("User repository initialized\n")

	// Создаём gRPC клиент для файлового сервиса
	fmt.Printf("Creating File Service client connection to %s:%d...\n", cfg.FileService.Host, cfg.FileService.Port)
	var fileServiceClient fileClient.FileServiceClient
	realClient, err := fileClient.NewFileServiceClient(cfg.FileService.Host, cfg.FileService.Port)
	if err != nil {
		fmt.Printf("⚠Failed to create file service client, using mock client for testing: %v\n", err)
		logBase.Info(ctx, "Failed to create file service client, using mock client for testing", zap.Error(err))

		// Используем mock клиент для тестирования
		// В продакшене здесь можно установить fileServiceClient = nil, чтобы регистрация не работала
		useMock := os.Getenv("USE_MOCK_FILE_SERVICE") == "true"
		if useMock {
			fileServiceClient = fileClient.NewMockFileServiceClient(false) // false = не должен падать
			fmt.Printf("Using mock file service client for testing\n")
			logBase.Info(ctx, "Using mock file service client for testing")
		} else {
			fileServiceClient = nil
			fmt.Printf("File service is required for registration. Set USE_MOCK_FILE_SERVICE=true for testing.\n")
			logBase.Info(ctx, "File service is required for registration. Set USE_MOCK_FILE_SERVICE=true for testing.")
		}
	} else {
		fileServiceClient = realClient
		fmt.Printf("File Service client created successfully\n")
		logBase.Info(ctx, "File service client created successfully")
	}

	// Создаём security
	fmt.Printf("Initializing security service...\n")
	securityService := security.NewSecurity(
		cfg.Jwt.SecretKey,
		cfg.Jwt.Expiration,
		cfg.Verification.SecretKey,
		cfg.Verification.Expiration,
	)
	fmt.Printf("Security service initialized\n")

	// Создаём сервис пользователей
	fmt.Printf("Initializing user service...\n")
	userService := service.NewUserService(userRepo, securityService, fileServiceClient)
	fmt.Printf("User service initialized\n")

	// Создаём gRPC сервер (фоново, ошибки логируем, но не блокируем HTTP)
	fmt.Printf("Starting gRPC auth server on port %d...\n", cfg.Grpc.Port)
	grpcSrv := authServer.NewAuthServer(&ctx, userService, securityService, &cfg.Grpc)
	go func() {
		if err := grpcSrv.StartAuthServer(); err != nil {
			logBase.Error(ctx, "Failed to start gRPC server", zap.Error(err))
		}
	}()

	// Создаём HTTP хэндлер и роутер
	fmt.Printf("Setting up HTTP handlers and routes...\n")
	handler := api.NewHandler(userService)
	router := api.SetupRoutes(handler)
	fmt.Printf("HTTP handlers and routes configured\n")

	// HTTP-сервер
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	fmt.Printf("Starting HTTP server on %s...\n", addr)
	logBase.Info(ctx, "Starting auth service", zap.String("http_addr", addr), zap.Int("grpc_port", cfg.Grpc.Port))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logBase.Error(ctx, "HTTP ListenAndServe failed", zap.Error(err))
		}
	}()

	fmt.Printf("Auth service started successfully!\n")
	fmt.Printf("HTTP Server: http://%s\n", addr)
	fmt.Printf("gRPC Server: localhost:%d\n", cfg.Grpc.Port)
	fmt.Println("==================================================")

	return srv, logBase, nil
}
