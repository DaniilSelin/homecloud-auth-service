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

	fmt.Printf("DBMANAGER CONNECT: %s:%d\n", cfg.DbManager.Host, cfg.DbManager.Port)
	// Создаём gRPC dbClient
	dbClient, err := dbClient.NewDBServiceClient(cfg.DbManager.Host, cfg.DbManager.Port)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dbClient: %w", err)
	}
	// Создаём репозиторий
	userRepo := repository.NewUserRepository(dbClient)

	// Создаём gRPC клиент для файлового сервиса
	fmt.Printf("FILE SERVICE CONNECT: %s:%d\n", cfg.FileService.Host, cfg.FileService.Port)
	fileServiceClient, err := fileClient.NewFileServiceClient(cfg.FileService.Host, cfg.FileService.Port)
	if err != nil {
		logBase.Info(ctx, "Failed to create file service client, continuing without file service", zap.Error(err))
		fileServiceClient = nil
	} else {
		logBase.Info(ctx, "File service client created successfully")
	}

	// Создаём security
	securityService := security.NewSecurity(
		cfg.Jwt.SecretKey,
		cfg.Jwt.Expiration,
		cfg.Verification.SecretKey,
		cfg.Verification.Expiration,
	)

	// Создаём сервис пользователей
	userService := service.NewUserService(userRepo, securityService, fileServiceClient)

	// Создаём gRPC сервер (фоново, ошибки логируем, но не блокируем HTTP)
	grpcSrv := authServer.NewAuthServer(&ctx, userService, securityService, &cfg.Grpc)
	go func() {
		if err := grpcSrv.StartAuthServer(); err != nil {
			logBase.Error(ctx, "Failed to start gRPC server", zap.Error(err))
		}
	}()

	// Создаём HTTP хэндлер и роутер
	handler := api.NewHandler(userService)
	router := api.SetupRoutes(handler)

	// HTTP-сервер
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	logBase.Info(ctx, "Starting auth service", zap.String("http_addr", addr), zap.Int("grpc_port", cfg.Grpc.Port))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logBase.Error(ctx, "HTTP ListenAndServe failed", zap.Error(err))
		}
	}()

	return srv, logBase, nil
}
