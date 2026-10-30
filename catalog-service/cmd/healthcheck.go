package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	catalogv1 "github.com/Holarbtw/homestock-backend/pkg/pb/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func checkHealth() error {
	port := os.Getenv("CATALOG_SERVICE_PORT")
	if port == "" {
		port = "9090"
	}
	conn, err := grpc.NewClient(net.JoinHostPort("127.0.0.1", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	response, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{
		Service: catalogv1.CatalogService_ServiceDesc.ServiceName,
	})
	if err != nil {
		return fmt.Errorf("gRPC health check failed: %w", err)
	}
	if response.Status != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("catalog is not ready: %s", response.Status)
	}
	return nil
}
