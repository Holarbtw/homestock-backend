package domain

import (
	"net/http"

	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
)

var ErrInvalidDate = &apperrors.CustomError{
	Code:       "INVALID_DATE",
	StatusCode: http.StatusBadRequest,
	Msg:        "Invalid calendar date, expected YYYY-MM-DD with a year from 0001 to 9999",
}
