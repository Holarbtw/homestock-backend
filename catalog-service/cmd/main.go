package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Holarbtw/homestock-backend/catalog-service/internal/adapters/db/postgres"
	grpcAdapters "github.com/Holarbtw/homestock-backend/catalog-service/internal/adapters/grpc"
	"github.com/Holarbtw/homestock-backend/catalog-service/internal/core/config"
	"github.com/Holarbtw/homestock-backend/catalog-service/internal/core/services"
	"github.com/Holarbtw/homestock-backend/pkg/logger"
	catalogv1 "github.com/Holarbtw/homestock-backend/pkg/pb/catalog/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	grpcgo "google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "check the local gRPC server and database")
	flag.Parse()

	var err error
	if *healthcheck {
		err = checkHealth()
	} else {
		err = run()
	}
	if err != nil {
		slog.Error("application terminated with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	logger.InitLogger("catalog-service")

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

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1
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

	productRepo := postgres.NewProductRepository(pool)

	catalogService := services.NewCatalogService(productRepo)

	grpcServer := grpcgo.NewServer(
		grpcgo.UnaryInterceptor(grpcAdapters.Interceptor),
	)
	defer grpcServer.Stop()

	catalogServer := grpcAdapters.NewCatalogServer(catalogService)

	catalogv1.RegisterCatalogServiceServer(grpcServer, catalogServer)
	healthpb.RegisterHealthServer(grpcServer, grpcAdapters.NewHealthServer(pool.Ping))

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("gRPC server starting", "addr", listener.Addr())
		serverErrors <- grpcServer.Serve(listener)
	}()

	select {
	case err := <-serverErrors:
		if err != nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
		return nil
	case <-ctx.Done():
		stop()
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	shutdownDone := make(chan struct{})

	go func() {
		grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		slog.Info("gRPC server stopped gracefully")

	case <-shutdownCtx.Done():
		slog.Warn("graceful shutdown timed out, stopping active RPCs")
		grpcServer.Stop()
	}

	return nil
}
