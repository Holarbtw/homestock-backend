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
)
