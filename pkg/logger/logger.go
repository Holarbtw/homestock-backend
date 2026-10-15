package logger

import (
	"log/slog"
	"os"
)

func InitLogger(serviceName string) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs([]slog.Attr{
		slog.String("service", serviceName),
	})

	slog.SetDefault(slog.New(handler))
}
