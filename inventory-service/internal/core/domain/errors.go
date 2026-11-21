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
		Code:       "INVALID_ITEM_UPDATE",
		StatusCode: http.StatusBadRequest,
		Msg:        "Invalid item update data",
	}
	ErrInvalidItemCategory = &apperrors.CustomError{
		Code:       "INVALID_ITEM_CATEGORY",
		StatusCode: http.StatusBadRequest,
		Msg:        "Invalid item category",
	}
	ErrInvalidItemName = &apperrors.CustomError{
		Code:       "INVALID_ITEM_NAME",
		StatusCode: http.StatusBadRequest,
		Msg:        "Item name must be nonblank and contain no NUL characters",
	}
	ErrItemNameTooLong = &apperrors.CustomError{
		Code:       "ITEM_NAME_TOO_LONG",
		StatusCode: http.StatusBadRequest,
		Msg:        fmt.Sprintf("Item name must not exceed %d characters", MaxItemNameLength),
	}
	ErrInvalidItemBrand = &apperrors.CustomError{
		Code:       "INVALID_ITEM_BRAND",
		StatusCode: http.StatusBadRequest,
		Msg:        "Item brand must be nonblank and contain no NUL characters",
	}
	ErrItemBrandTooLong = &apperrors.CustomError{
		Code:       "ITEM_BRAND_TOO_LONG",
		StatusCode: http.StatusBadRequest,
		Msg:        fmt.Sprintf("Item brand must not exceed %d characters", MaxItemBrandLength),
	}
	ErrInvalidItemCreate = &apperrors.CustomError{
		Code:       "INVALID_ITEM_CREATE",
		StatusCode: http.StatusBadRequest,
		Msg:        "Invalid item creation data",
	}
	ErrInvalidItemUnit = &apperrors.CustomError{
		Code:       "INVALID_ITEM_UNIT",
		StatusCode: http.StatusBadRequest,
		Msg:        "Item unit must be ml, g or pcs",
	}
	ErrInvalidGTIN = &apperrors.CustomError{
		Code:       "INVALID_GTIN",
		StatusCode: http.StatusBadRequest,
		Msg:        "GTIN must contain 8 or 12 to 14 digits",
	}
	ErrInvalidCatalogProductID = &apperrors.CustomError{
		Code:       "INVALID_CATALOG_PRODUCT_ID",
		StatusCode: http.StatusBadRequest,
		Msg:        "Catalog product ID must be a nonzero UUID",
	}
	ErrInvalidPackageQuantity = &apperrors.CustomError{
		Code:       "INVALID_PACKAGE_QUANTITY",
		StatusCode: http.StatusBadRequest,
		Msg:        "Package quantity must be greater than zero",
	}
	ErrInvalidExternalImageURL = &apperrors.CustomError{
		Code:       "INVALID_EXTERNAL_IMAGE_URL",
		StatusCode: http.StatusBadRequest,
		Msg:        "External image URL must be an absolute HTTP or HTTPS URL with a hostname",
	}
	ErrInvalidBatchQuantity = &apperrors.CustomError{
		Code:       "INVALID_BATCH_QUANTITY",
		StatusCode: http.StatusBadRequest,
		Msg:        "Batch quantity must be greater than zero",
	}
	ErrPurchaseDateInFuture = &apperrors.CustomError{
		Code:       "PURCHASE_DATE_IN_FUTURE",
		StatusCode: http.StatusBadRequest,
		Msg:        "Purchase date must not be in the future",
	}
	ErrInvalidIdempotencyKey = &apperrors.CustomError{
		Code:       "INVALID_IDEMPOTENCY_KEY",
		StatusCode: http.StatusBadRequest,
		Msg:        fmt.Sprintf("Idempotency key must be nonblank UTF-8 text of at most %d bytes without NUL characters", MaxIdempotencyKeyBytes),
	}
	ErrIdempotencyKeyConflict = &apperrors.CustomError{
		Code:       "IDEMPOTENCY_KEY_CONFLICT",
		StatusCode: http.StatusConflict,
		Msg:        "Idempotency key was already used for a different request",
	}
)
