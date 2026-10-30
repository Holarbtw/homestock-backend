package grpc

import (
	"context"
	"time"

	catalogv1 "github.com/Holarbtw/homestock-backend/pkg/pb/catalog/v1"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type HealthServer struct {
	healthpb.UnimplementedHealthServer
	ping func(context.Context) error
}

func NewHealthServer(ping func(context.Context) error) *HealthServer {
	return &HealthServer{ping: ping}
}

func (s *HealthServer) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	if req.Service != "" && req.Service != catalogv1.CatalogService_ServiceDesc.ServiceName {
		return nil, status.Error(codes.NotFound, "Unknown service")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	state := healthpb.HealthCheckResponse_SERVING
	if err := s.ping(ctx); err != nil {
		state = healthpb.HealthCheckResponse_NOT_SERVING
	}
	return &healthpb.HealthCheckResponse{Status: state}, nil
}
