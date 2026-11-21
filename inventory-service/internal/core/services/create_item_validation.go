package services

import (
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/google/uuid"
)

// Match the catalog's accepted GTIN lengths before padding to 14 digits.
var itemGTINPattern = regexp.MustCompile(`^([0-9]{8}|[0-9]{12,14})$`)

// validateCreateItemInput returns normalized input for fingerprinting and persistence.
// The location of now determines the calendar date used for purchase validation.
func validateCreateItemInput(input *ports.CreateItemInput, idempotencyKey string, now time.Time) (*ports.CreateItemInput, error) {
	if input == nil {
		return nil, domain.ErrInvalidItemCreate
	}

	name, nameErr := validateItemName(input.Name)
	category, categoryErr := validateItemCategory(string(input.Category))
	unit, unitErr := validateItemUnit(input.Unit)
	brand, brandErr := normalizeOptionalItemBrand(input.Brand)
	gtin, gtinErr := normalizeItemGTIN(input.GTIN)

	type fieldCheck struct {
		field string
		err   error
	}
	checks := []fieldCheck{
		{"name", nameErr},
		{"brand", brandErr},
		{"category", categoryErr},
		{"unit", unitErr},
		{"gtin", gtinErr},
		{"external_image_url", validateExternalImageURL(input.ExternalImageURL)},
		{"idempotency_key", validateIdempotencyKey(idempotencyKey)},
	}
	if input.CatalogProductID != nil && *input.CatalogProductID == uuid.Nil {
		checks = append(checks, fieldCheck{"catalog_product_id", domain.ErrInvalidCatalogProductID})
	}
	if input.PackageQuantity != nil && input.PackageQuantity.Thousandths() <= 0 {
		checks = append(checks, fieldCheck{"package_quantity", domain.ErrInvalidPackageQuantity})
	}
	if firstBatch := input.FirstBatch; firstBatch != nil {
		if firstBatch.Quantity.Thousandths() <= 0 {
			checks = append(checks, fieldCheck{"first_batch.quantity", domain.ErrInvalidBatchQuantity})
		}
		checks = append(checks,
			fieldCheck{"first_batch.purchased_on", validatePurchaseDate(firstBatch.PurchasedOn, now)},
			fieldCheck{"first_batch.expires_on", validateOptionalDate(firstBatch.ExpiresOn)},
		)
	}

	var fields []apperrors.FieldError
	for _, check := range checks {
		var err error
		fields, err = appendFieldError(fields, check.field, check.err)
		if err != nil {
			return nil, err
		}
	}
	if len(fields) > 0 {
		return nil, domain.ErrInvalidItemCreate.WithFields(fields)
	}

	normalized := *input
	normalized.Name = name
	normalized.Brand = brand
	normalized.Category = category
	normalized.Unit = unit
	normalized.GTIN = gtin
	return &normalized, nil
}

func validateItemUnit(value domain.Unit) (domain.Unit, error) {
	unit := domain.Unit(strings.TrimSpace(string(value)))
	switch unit {
	case domain.Milliliter, domain.Gram, domain.Pcs:
		return unit, nil
	default:
		return "", domain.ErrInvalidItemUnit
	}
}

func normalizeOptionalItemBrand(brand *string) (*string, error) {
	if brand == nil {
		return nil, nil
	}
	normalized, err := validateItemBrand(*brand)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func normalizeItemGTIN(gtin *string) (*string, error) {
	if gtin == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*gtin)
	if !itemGTINPattern.MatchString(trimmed) {
		return nil, domain.ErrInvalidGTIN
	}
	normalized := strings.Repeat("0", 14-len(trimmed)) + trimmed
	return &normalized, nil
}

func validateExternalImageURL(imageURL *url.URL) error {
	if imageURL == nil {
		return nil
	}
	if (imageURL.Scheme != "http" && imageURL.Scheme != "https") || imageURL.Hostname() == "" {
		return domain.ErrInvalidExternalImageURL
	}
	return nil
}

func validateIdempotencyKey(key string) error {
	if len(key) > domain.MaxIdempotencyKeyBytes || strings.TrimSpace(key) == "" ||
		!utf8.ValidString(key) || strings.ContainsRune(key, '\x00') {
		return domain.ErrInvalidIdempotencyKey
	}
	return nil
}

func validateOptionalDate(date *domain.Date) error {
	if date != nil && date.Year() == 0 {
		return domain.ErrInvalidDate
	}
	return nil
}

func validatePurchaseDate(date *domain.Date, now time.Time) error {
	if err := validateOptionalDate(date); err != nil {
		return err
	}
	if date != nil && date.String() > now.Format(time.DateOnly) {
		return domain.ErrPurchaseDateInFuture
	}
	return nil
}
