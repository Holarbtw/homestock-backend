package domain

import (
	"net/http"

	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
)

var (
	ErrInvalidDate = &apperrors.CustomError{
		Code:       "INVALID_DATE",
		StatusCode: http.StatusBadRequest,
		Msg:        "Invalid calendar date, expected YYYY-MM-DD with a year from 0001 to 9999",
	}

	ErrItemNotFound = &apperrors.CustomError{
		Code:       "ITEM_NOT_FOUND",
		StatusCode: http.StatusNotFound,
		Msg:        "Inventory item not found",
	}

	ErrSpaceNotFound = &apperrors.CustomError{
		Code:       "SPACE_NOT_FOUND",
		StatusCode: http.StatusNotFound,
		Msg:        "Inventory space not found",
	}
)
