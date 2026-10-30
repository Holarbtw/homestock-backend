package domain

import (
	"net/http"

	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
)

var ErrInvalidGTIN = &apperrors.CustomError{
	Code:       "INVALID_GTIN",
	StatusCode: http.StatusBadRequest,
	Msg:        "GTIN must contain 8, 12, 13, or 14 digits",
}
