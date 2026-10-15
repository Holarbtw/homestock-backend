package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Holarbtw/homestock-backend/auth-service/internal/adapters/db/postgres"
	"github.com/Holarbtw/homestock-backend/auth-service/internal/config"
	"github.com/Holarbtw/homestock-backend/pkg/logger"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

	_ = pkgPostgres.NewTxManager(pool)
	_ = postgres.NewUserRepository(pool)

	slog.Info("data layer initialized successfully")
	return nil
}
