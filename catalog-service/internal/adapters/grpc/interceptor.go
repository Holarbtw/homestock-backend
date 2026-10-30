package grpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Holarbtw/homestock-backend/catalog-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func Interceptor(
	ctx context.Context,
	req any,
	info *grpcgo.UnaryServerInfo,
	handler grpcgo.UnaryHandler,
) (any, error) {
	if info.FullMethod == healthpb.Health_Check_FullMethodName {
		return handler(ctx, req)
	}
	start := time.Now()

	response, err := handler(ctx, req)
	duration := time.Since(start)

	if err == nil {

		attrs := []any{
			"method", info.FullMethod,
			"duration", duration,
		}

		slog.InfoContext(ctx, "request processed", attrs...)
		return response, nil
	}

	attrs := []any{
		"method", info.FullMethod,
		"duration", duration,
		"err", err,
	}

	grpcErr, level := toGRPCError(err)
	slog.Log(ctx, level, "request error", attrs...)

	return nil, grpcErr
}

func toGRPCError(err error) (error, slog.Level) {
	if err == nil {
		return nil, slog.LevelInfo
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "Request canceled"), slog.LevelInfo
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "Request deadline exceeded"), slog.LevelWarn
	}

	var appErr *apperrors.CustomError

	if errors.As(err, &appErr) {
		switch appErr.Code {
		case domain.ErrInvalidGTIN.Code:
			return status.Error(codes.InvalidArgument, appErr.Msg), slog.LevelWarn
		case apperrors.ErrNotFound.Code:
			return status.Error(codes.NotFound, appErr.Msg), slog.LevelWarn
		}
	}

	return status.Error(codes.Internal, "Internal server error"), slog.LevelError
}
