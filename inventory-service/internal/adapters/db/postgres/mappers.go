package postgres

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/url"
	"time"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/jackc/pgx/v5/pgtype"
)

func toIdempotencyRecord(record *idempotencyRecordDB) (*ports.IdempotencyRecord, error) {
	if record == nil {
		return nil, nil
	}
	if len(record.RequestHash) != sha256.Size {
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("persisted idempotency hash has length %d, expected %d", len(record.RequestHash), sha256.Size),
		)
	}

	result := ports.IdempotencyRecord{
		SpaceID:     record.SpaceID,
		RequestKey:  record.RequestKey,
		CommandType: record.CommandType,
		CreatedAt:   record.CreatedAt,
	}
	copy(result.RequestHash[:], record.RequestHash)
	if record.ResultData != nil {
		result.Result = &ports.IdempotencyResult{
			Data: bytes.Clone(record.ResultData),
		}
	}

	return &result, nil
}

func toDomainSpace(record *spaceDB) *domain.Space {
	if record == nil {
		return nil
	}
	return &domain.Space{
		ID:          record.ID,
		OwnerUserID: record.OwnerUserID,
		CreatedAt:   record.CreatedAt,
	}
}

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

func toDomainItemStock(record *itemStockDB) (domain.ItemStock, error) {
	item, err := toDomainItem(&record.itemDB)
	if err != nil {
		return domain.ItemStock{}, err
	}
	total, err := quantity.ParseQuantity(record.TotalQuantity)
	if err != nil {
		return domain.ItemStock{}, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("parse total batch quantity for item %s: %w", item.ID, err),
		)
	}
	return domain.ItemStock{Item: *item, TotalQuantity: *total}, nil
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

func toDBBatch(batch domain.Batch) (batchDB, error) {
	purchasedOn, err := toDBDate(batch.PurchasedOn)
	if err != nil {
		return batchDB{}, err
	}
	expiresOn, err := toDBDate(batch.ExpiresOn)
	if err != nil {
		return batchDB{}, err
	}
	return batchDB{
		SpaceID:           batch.SpaceID,
		ItemID:            batch.ItemID,
		RemainingQuantity: batch.RemainingQuantity.String(),
		PurchasedOn:       purchasedOn,
		ExpiresOn:         expiresOn,
	}, nil
}

func toDomainBatch(record *batchDB) (*domain.Batch, error) {
	remaining, err := quantity.ParseQuantity(record.RemainingQuantity)
	if err != nil {
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("parse remaining quantity for batch %s: %w", record.ID, err),
		)
	}
	purchasedOn, err := toDomainDate(record.PurchasedOn)
	if err != nil {
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("parse purchase date for batch %s: %w", record.ID, err),
		)
	}
	expiresOn, err := toDomainDate(record.ExpiresOn)
	if err != nil {
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("parse expiry date for batch %s: %w", record.ID, err),
		)
	}
	return &domain.Batch{
		ID:                record.ID,
		SpaceID:           record.SpaceID,
		ItemID:            record.ItemID,
		RemainingQuantity: *remaining,
		PurchasedOn:       purchasedOn,
		AddedAt:           record.AddedAt,
		ExpiresOn:         expiresOn,
	}, nil
}

func toDBDate(date *domain.Date) (pgtype.Date, error) {
	if date == nil {
		return pgtype.Date{}, nil
	}
	if date.Year() == 0 {
		return pgtype.Date{}, domain.ErrInvalidDate
	}
	return pgtype.Date{
		Time:  time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
		Valid: true,
	}, nil
}

func toDomainDate(date pgtype.Date) (*domain.Date, error) {
	if !date.Valid {
		return nil, nil
	}
	if date.InfinityModifier != pgtype.Finite {
		return nil, domain.ErrInvalidDate
	}
	return domain.ParseDate(date.Time.Format(time.DateOnly))
}

func toDomainStockOperation(record *stockOperationDB) (*domain.StockOperation, error) {
	operationType := domain.StockOperationType(record.Type)
	switch operationType {
	case domain.Replenish, domain.Consume, domain.WriteOff:
	default:
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("stock operation %s has an invalid persisted type: %q", record.ID, record.Type),
		)
	}
	amount, err := quantity.ParseQuantity(record.Quantity)
	if err != nil {
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("parse quantity for operation %s: %w", record.ID, err),
		)
	}
	return &domain.StockOperation{
		ID:          record.ID,
		SpaceID:     record.SpaceID,
		ItemID:      record.ItemID,
		ActorUserID: record.ActorUserID,
		Type:        operationType,
		Quantity:    *amount,
		OccurredAt:  record.OccurredAt,
	}, nil
}
