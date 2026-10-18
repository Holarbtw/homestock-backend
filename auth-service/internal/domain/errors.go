package domain

import (
	"net/http"

	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
)

var (
	ErrInvalidCredentials = &apperrors.CustomError{
		Code:       "INVALID_CREDENTIALS",
		StatusCode: http.StatusUnauthorized,
		Msg:        "Invalid email or password",
	}

	ErrEmailAlreadyExists = &apperrors.CustomError{
		Code:       "EMAIL_ALREADY_EXISTS",
		StatusCode: http.StatusConflict,
		Msg:        "User with this email already exists",
	}

	ErrSessionRevoked = &apperrors.CustomError{
		Code:       "SESSION_REVOKED",
		StatusCode: http.StatusUnauthorized,
		Msg:        "Session has been revoked or expired",
	}

	ErrGuestAlreadyLinked = &apperrors.CustomError{
		Code:       "GUEST_ALREADY_LINKED",
		StatusCode: http.StatusBadRequest,
		Msg:        "Account is already linked to email",
	}

	ErrInvalidTimezone = &apperrors.CustomError{
		Code:       "INVALID_TIMEZONE",
		StatusCode: http.StatusBadRequest,
		Msg:        "Invalid timezone format, expected IANA format (e.g. Europe/Moscow)",
	}

	ErrInvalidEmail = &apperrors.CustomError{
		Code:       "INVALID_EMAIL",
		StatusCode: http.StatusBadRequest,
		Msg:        "Invalid email address format",
	}

	ErrInvalidPassword = &apperrors.CustomError{
		Code:       "INVALID_PASSWORD",
		StatusCode: http.StatusBadRequest,
		Msg:        "Password must be between 8 and 72 characters",
	}

	ErrInvalidNickname = &apperrors.CustomError{
		Code:       "INVALID_NICKNAME",
		StatusCode: http.StatusBadRequest,
		Msg:        "Nickname must be between 2 and 50 characters",
	}
)
