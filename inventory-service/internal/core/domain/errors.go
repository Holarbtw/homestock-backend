package domain

import (
	"fmt"
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
	ErrInvalidItemUpdate = &apperrors.CustomError{
		Code: "INVALID_ITEM_UPDATE",
		StatusCode: http.StatusBadRequest,
		Msg: "Invalid item update data",
	}
	ErrInvalidItemCategory = &apperrors.CustomError{
		Code: "INVALID_ITEM_CATEGORY",
		StatusCode: http.StatusBadRequest,
		Msg: "Invalid item category",
	}
	ErrInvalidItemName = &apperrors.CustomError{
		Code: "INVALID_ITEM_NAME",
		StatusCode: http.StatusBadRequest,
		Msg: "Invalid item name",
	}
	ErrItemNameTooLong = &apperrors.CustomError{
		Code: "ITEM_NAME_TOO_LONG",
		StatusCode: http.StatusBadRequest,
		Msg: fmt.Sprintf("Item name must not exceed %d characters", MaxItemNameLength),
	}
	ErrInvalidItemBrand = &apperrors.CustomError{
		Code: "INVALID_ITEM_BRAND",
		StatusCode: http.StatusBadRequest,
		Msg: "Invalid item brand",
	}
	ErrItemBrandTooLong = &apperrors.CustomError{
		Code: "ITEM_BRAND_TOO_LONG",
		StatusCode: http.StatusBadRequest,
		Msg: fmt.Sprintf("Item brand must not exceed %d characters", MaxItemBrandLength),
	}

)
