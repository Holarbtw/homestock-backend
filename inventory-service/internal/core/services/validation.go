package services

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
)

func validateItemCategory(category string) (domain.Category, error) {
	trimmed := strings.TrimSpace(category)
	switch trimmed {
	case string(domain.Food):
		return domain.Food, nil
	case string(domain.Household):
		return domain.Household, nil
	case string(domain.Medicine):
		return domain.Medicine, nil
	default:
		return "", domain.ErrInvalidItemCategory
	}
}

func validateItemName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || strings.ContainsRune(trimmed, '\x00') {
		return "", domain.ErrInvalidItemName
	}

	if utf8.RuneCountInString(trimmed) > domain.MaxItemNameLength {
		return "", domain.ErrItemNameTooLong
	}

	return trimmed, nil
}

func validateItemBrand(brand string) (string, error) {
	trimmed := strings.TrimSpace(brand)
	if trimmed == "" || strings.ContainsRune(trimmed, '\x00') {
		return "", domain.ErrInvalidItemBrand
	}

	if utf8.RuneCountInString(trimmed) > domain.MaxItemBrandLength {
		return "", domain.ErrItemBrandTooLong
	}

	return trimmed, nil
}

func validateAddBatchInput(input *ports.AddBatchInput) error{
	if input == nil{
		return apperrors.ErrInvalidRequest
	}
	if input.Quantity.Thousandths() == 0{
		return domain.ErrInvalidBatchQuantity
	}

	err := validatePurchaseDate(input.PurchasedOn, time.Now())
	if err != nil {
		return err
	}

	return validateOptionalDate(input.ExpiresOn)
}

func appendFieldError(
	fields []apperrors.FieldError,
	field string,
	err error,
) ([]apperrors.FieldError, error) {
	if err == nil {
		return fields, nil
	}

	var customErr *apperrors.CustomError
	if !errors.As(err, &customErr) {
		return fields, err
	}

	fields = append(fields, apperrors.FieldError{
		Field: field,
		Code:  customErr.Code,
	})

	return fields, nil
}
