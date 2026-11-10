package postgres

import (
	"fmt"
	"net/url"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
)

func toDomainItem(record *itemDB) (*domain.InventoryItem, error) {
	if record == nil {
		return nil, nil
	}

	item := domain.InventoryItem{
		ID:               record.ID,
		SpaceID:          record.SpaceID,
		CatalogProductID: record.CatalogProductID,
		GTIN:             record.GTIN,
		Name:             record.Name,
		Brand:            record.Brand,
		Category:         domain.Category(record.Category),
		Unit:             domain.Unit(record.Unit),
		ImageObjectKey:   record.ImageObjectKey,
		CreatedAt:        record.CreatedAt,
		UpdatedAt:        record.UpdatedAt,
		DeletedAt:        record.DeletedAt,
	}

	switch item.Category {
	case domain.Food, domain.Medicine, domain.Household:
	default:
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("inventory item %s has an invalid persisted category: %q", item.ID, record.Category),
		)
	}

	switch item.Unit {
	case domain.Milliliter, domain.Gram, domain.Pcs:
	default:
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("inventory item %s has an invalid persisted unit: %q", item.ID, record.Unit),
		)
	}

	if record.MinimumQuantity != nil {
		parsed, err := quantity.ParseQuantity(*record.MinimumQuantity)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("parse inventory item %s minimum quantity: %w", item.ID, err),
			)
		}
		item.MinimumQuantity = parsed
	}

	if record.PackageQuantity != nil {
		parsed, err := quantity.ParseQuantity(*record.PackageQuantity)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("parse inventory item %s package quantity: %w", item.ID, err),
			)
		}
		if parsed.Thousandths() == 0 {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("inventory item %s has a zero persisted package quantity", item.ID),
			)
		}
		item.PackageQuantity = parsed
	}

	if record.ExternalImageURL != nil {
		parsed, err := url.Parse(*record.ExternalImageURL)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}
		if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("inventory item %s has an invalid persisted image URL", item.ID),
			)
		}
		if record.ImageObjectKey != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("inventory item %s has two persisted image sources", item.ID),
			)
		}
		item.ExternalImageURL = parsed
	}

	return &item, nil
}

func toDBItem(item domain.InventoryItem) itemDB {
	record := itemDB{
		ID:               item.ID,
		SpaceID:          item.SpaceID,
		CatalogProductID: item.CatalogProductID,
		GTIN:             item.GTIN,
		Name:             item.Name,
		Brand:            item.Brand,
		Category:         string(item.Category),
		Unit:             string(item.Unit),
		ImageObjectKey:   item.ImageObjectKey,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
		DeletedAt:        item.DeletedAt,
	}

	if item.MinimumQuantity != nil {
		value := item.MinimumQuantity.String()
		record.MinimumQuantity = &value
	}
	if item.PackageQuantity != nil {
		value := item.PackageQuantity.String()
		record.PackageQuantity = &value
	}
	if item.ExternalImageURL != nil {
		value := item.ExternalImageURL.String()
		record.ExternalImageURL = &value
	}

	return record
}
