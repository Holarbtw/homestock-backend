package apperrors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type errorContextKey struct{}

var errCtxKey = errorContextKey{}

type CustomError struct {
	Code       string `json:"code"`
	StatusCode int    `json:"status_code"`
	Msg        string `json:"message"`
	Err        error  `json:"-"`
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Msg, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Msg)
}

func (e *CustomError) Unwrap() error {
	return e.Err
}

func (e *CustomError) WithInternal(err error) *CustomError {
	return &CustomError{
		Code:       e.Code,
		StatusCode: e.StatusCode,
		Msg:        e.Msg,
		Err:        err,
	}
}

var (
	ErrInternalServer   = &CustomError{Code: "INTERNAL_SERVER_ERROR", StatusCode: http.StatusInternalServerError, Msg: "Internal server error"}
	ErrInvalidRequest   = &CustomError{Code: "INVALID_REQUEST", StatusCode: http.StatusBadRequest, Msg: "Invalid request payload"}
	ErrUnauthorized     = &CustomError{Code: "UNAUTHORIZED", StatusCode: http.StatusUnauthorized, Msg: "Unauthorized"}
	ErrInvalidToken     = &CustomError{Code: "INVALID_TOKEN", StatusCode: http.StatusUnauthorized, Msg: "Invalid or expired token"}
	ErrForbidden        = &CustomError{Code: "FORBIDDEN", StatusCode: http.StatusForbidden, Msg: "Access forbidden"}
	ErrNotFound         = &CustomError{Code: "NOT_FOUND", StatusCode: http.StatusNotFound, Msg: "Resource not found"}
	ErrConflict         = &CustomError{Code: "CONFLICT", StatusCode: http.StatusConflict, Msg: "Resource already exists"}
	ErrValidationFailed = &CustomError{Code: "VALIDATION_FAILED", StatusCode: http.StatusBadRequest, Msg: "Validation failed"}
)

func RespondWithError(w http.ResponseWriter, r *http.Request, err error) {
	*r = *r.WithContext(context.WithValue(r.Context(), errCtxKey, err))

	w.Header().Set("Content-Type", "application/json")

	var customErr *CustomError
	if errors.As(err, &customErr) {
		w.WriteHeader(customErr.StatusCode)
		_ = json.NewEncoder(w).Encode(customErr)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(ErrInternalServer)
}

func GetErrorFromContext(ctx context.Context) error {
	if err, ok := ctx.Value(errCtxKey).(error); ok {
		return err
	}
	return nil
}
