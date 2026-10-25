package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Holarbtw/homestock-backend/auth-service/docs"
	"github.com/Holarbtw/homestock-backend/auth-service/internal/adapters/db/postgres"
	httpAdapter "github.com/Holarbtw/homestock-backend/auth-service/internal/adapters/http"
	"github.com/Holarbtw/homestock-backend/auth-service/internal/adapters/security"
	"github.com/Holarbtw/homestock-backend/auth-service/internal/config"
	"github.com/Holarbtw/homestock-backend/auth-service/internal/core/services"
	pkgJWT "github.com/Holarbtw/homestock-backend/pkg/jwt"
	"github.com/Holarbtw/homestock-backend/pkg/logger"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @title Homestock Auth Service API
// @version 1.0
// @description Authentication and user profile service for Homestock project.
// @BasePath /api/v1/auth
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer " followed by a space and then your JWT access token.
func main() {
	if err := run(); err != nil {
		slog.Error("application terminated with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	logger.InitLogger("auth-service")

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	poolConfig, err := pgxpool.ParseConfig(cfg.DB.DSN())
	if err != nil {
		return fmt.Errorf("failed to parse db dsn: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create pgxpool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping postgres: %w", err)
	}
	slog.Info("connected to postgres", "db", cfg.DB.Name)

	txManager := pkgPostgres.NewTxManager(pool)
	userRepo := postgres.NewUserRepository(pool)
	hasher := security.NewBcryptHasher(cfg.Security.BcryptCost)
	tokenManager := security.NewJWTTokenManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL(),
		cfg.JWT.RefreshTokenTTL(),
	)

	authService := services.NewAuthService(userRepo, txManager, hasher, tokenManager)
	jwtVerifier := pkgJWT.NewTokenVerifier(cfg.JWT.Secret)

	authHandler := httpAdapter.NewAuthHandler(authService)
	router := httpAdapter.NewRouter(authHandler, jwtVerifier)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("http server error: %w", err)
		}
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received, closing server...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
		_ = server.Close()
		return fmt.Errorf("failed to gracefully shutdown server: %w", err)
	}

	slog.Info("server exited cleanly")
	return nil
}
