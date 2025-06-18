package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
	"homecloud-auth-service/config"
	"homecloud-auth-service/internal/logger"
	"homecloud-auth-service/internal/repository"
	"homecloud-auth-service/internal/security"
	"homecloud-auth-service/internal/service"
	"homecloud-auth-service/internal/transport/grpc/authServer"
	"homecloud-auth-service/internal/transport/http/api"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	srv, grpcSrv, logBase, err := run(ctx, os.Stdout, os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	<-ctx.Done()
	logBase.Info(ctx, "Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Gracefully stop both servers
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logBase.Error(ctx, "HTTP server shutdown failed", zap.Error(err))
	}
	if err := grpcSrv.Stop(); err != nil {
		logBase.Error(ctx, "gRPC server shutdown failed", zap.Error(err))
	}
	
	logBase.Info(ctx, "Servers exited gracefully")
}

func run(ctx context.Context, w io.Writer, args []string) (*http.Server, authServer.AuthServiceServer, *logger.Logger, error) {
	// Конфиг и логгер
	cfg, err := config.LoadConfig("config/config.local.yaml")
	if err != nil {
		return nil, nil, nil, err
	}
	logBase, err := logger.New(cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	ctx = logger.CtxWWithLogger(ctx, logBase)

	// Создаем репозиторий (заглушка для gRPC)
	userRepo := repository.NewUserRepository()

	// Создаем security
	securityService := security.NewSecurity(
		cfg.Jwt.SecretKey,
		cfg.Jwt.Expiration,
		cfg.Verification.SecretKey,
		cfg.Verification.Expiration,
	)

	// Создаем сервис пользователей
	userService := service.NewUserService(userRepo, securityService)

	// Создаем gRPC сервер
	grpcSrv := authServer.NewAuthServiceServer(userService, securityService, logBase, cfg.Grpc.Port)
	go func() {
		if err := grpcSrv.Start(); err != nil {
			logBase.Error(ctx, "Failed to start gRPC server", zap.Error(err))
		}
	}()

	// Создаем HTTP хэндлер и роутер
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

	return srv, grpcSrv, logBase, nil
}