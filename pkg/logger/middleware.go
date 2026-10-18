package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/go-chi/chi/v5/middleware"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		requestID := middleware.GetReqID(req.Context())
		if requestID != "" {
			w.Header().Set("X-Request-ID", requestID)
		}

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, req)

		duration := time.Since(start)

		attrs := []any{
			"request_id", requestID,
			"method", req.Method,
			"path", req.URL.Path,
			"status", wrapped.statusCode,
			"duration", duration,
			"size", wrapped.size,
		}

		if appErr := apperrors.GetErrorFromContext(req.Context()); appErr != nil {
			attrs = append(attrs, "err", appErr)
		}

		switch {
		case wrapped.statusCode >= http.StatusInternalServerError:
			slog.Error("request failed", attrs...)
		case wrapped.statusCode >= http.StatusBadRequest:
			slog.Warn("client request error", attrs...)
		default:
			slog.Info("request processed", attrs...)
		}
	})
}
