package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

type storedItemResult struct {
	Item      storedItemStock       `json:"item"`
	Batches   []storedBatch         `json:"batches"`
	Operation *storedStockOperation `json:"operation,omitempty"`
}

type storedItemStock struct {
	ID               uuid.UUID       `json:"id"`
	SpaceID          uuid.UUID       `json:"space_id"`
	CatalogProductID *uuid.UUID      `json:"catalog_product_id,omitempty"`
	GTIN             *string         `json:"gtin,omitempty"`
	Name             string          `json:"name"`
	Brand            *string         `json:"brand,omitempty"`
	Category         domain.Category `json:"category"`
	Unit             domain.Unit     `json:"unit"`
	MinimumQuantity  *string         `json:"minimum_quantity,omitempty"`
	PackageQuantity  *string         `json:"package_quantity,omitempty"`
	ExternalImageURL *string         `json:"external_image_url,omitempty"`
	ImageObjectKey   *string         `json:"image_object_key,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        *time.Time      `json:"deleted_at,omitempty"`
	TotalQuantity    string          `json:"total_quantity"`
}

type storedBatch struct {
	ID                uuid.UUID `json:"id"`
	SpaceID           uuid.UUID `json:"space_id"`
	ItemID            uuid.UUID `json:"item_id"`
	RemainingQuantity string    `json:"remaining_quantity"`
	PurchasedOn       *string   `json:"purchased_on,omitempty"`
	AddedAt           time.Time `json:"added_at"`
	ExpiresOn         *string   `json:"expires_on,omitempty"`
}

type storedStockOperation struct {
	ID          uuid.UUID                  `json:"id"`
	SpaceID     uuid.UUID                  `json:"space_id"`
	ItemID      uuid.UUID                  `json:"item_id"`
	ActorUserID uuid.UUID                  `json:"actor_user_id"`
	Type        domain.StockOperationType  `json:"type"`
	Quantity    string                     `json:"quantity"`
	OccurredAt  time.Time                  `json:"occurred_at"`
	Lines       []storedStockOperationLine `json:"lines"`
}

type storedStockOperationLine struct {
	SpaceID     uuid.UUID `json:"space_id"`
	OperationID uuid.UUID `json:"operation_id"`
	ItemID      uuid.UUID `json:"item_id"`
	BatchID     uuid.UUID `json:"batch_id"`
	Quantity    string    `json:"quantity"`
}

func encodeItemResult(stock *domain.ItemStock, batches []domain.Batch, operation *domain.StockOperation) ([]byte, error) {
	if stock == nil || stock.Item.ID == uuid.Nil {
		return nil, apperrors.ErrInternalServer.WithInternal(errors.New("create item result is missing the created item"))
	}
	result := storedItemResult{
		Item:      snapshotItemStock(stock),
		Batches:   make([]storedBatch, 0, len(batches)),
		Operation: snapshotStockOperation(operation),
	}
	for _, batch := range batches {
		result.Batches = append(result.Batches, storedBatch{
			ID:                batch.ID,
			SpaceID:           batch.SpaceID,
			ItemID:            batch.ItemID,
			RemainingQuantity: batch.RemainingQuantity.String(),
			PurchasedOn:       snapshotDate(batch.PurchasedOn),
			AddedAt:           batch.AddedAt,
			ExpiresOn:         snapshotDate(batch.ExpiresOn),
		})
	}
	data, err := json.Marshal(result)
	if err != nil {
		return nil, apperrors.ErrInternalServer.WithInternal(fmt.Errorf("encode create item result: %w", err))
	}
	return data, nil
}

func decodeItemResult(data []byte) (*domain.ItemStock, []domain.Batch, *domain.StockOperation, error) {
	var result storedItemResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, nil, nil, apperrors.ErrInternalServer.WithInternal(fmt.Errorf("decode create item result: %w", err))
	}
	stock, err := result.Item.toDomain()
	if err != nil {
		return nil, nil, nil, apperrors.ErrInternalServer.WithInternal(fmt.Errorf("decode create item result item: %w", err))
	}
	batches := make([]domain.Batch, 0, len(result.Batches))
	for i, batch := range result.Batches {
		parsed, err := batch.toDomain()
		if err != nil {
			return nil, nil, nil, apperrors.ErrInternalServer.WithInternal(fmt.Errorf("decode create item result batch %d: %w", i, err))
		}
		batches = append(batches, parsed)
	}
	operation, err := result.Operation.toDomain()
	if err != nil {
		return nil, nil, nil, apperrors.ErrInternalServer.WithInternal(fmt.Errorf("decode create item result operation: %w", err))
	}
	return stock, batches, operation, nil
}

func snapshotItemStock(stock *domain.ItemStock) storedItemStock {
	item := stock.Item
	result := storedItemStock{
		ID:               item.ID,
		SpaceID:          item.SpaceID,
		CatalogProductID: item.CatalogProductID,
		GTIN:             item.GTIN,
		Name:             item.Name,
		Brand:            item.Brand,
		Category:         item.Category,
		Unit:             item.Unit,
		MinimumQuantity:  snapshotQuantity(item.MinimumQuantity),
		PackageQuantity:  snapshotQuantity(item.PackageQuantity),
		ImageObjectKey:   item.ImageObjectKey,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
		DeletedAt:        item.DeletedAt,
		TotalQuantity:    stock.TotalQuantity.String(),
	}
	if item.ExternalImageURL != nil {
		text := item.ExternalImageURL.String()
		result.ExternalImageURL = &text
	}
	return result
}

func (item storedItemStock) toDomain() (*domain.ItemStock, error) {
	if item.ID == uuid.Nil {
		return nil, errors.New("missing item ID")
	}
	total, err := quantity.ParseQuantity(item.TotalQuantity)
	if err != nil {
		return nil, fmt.Errorf("total_quantity: %w", err)
	}
	minimum, err := parseSnapshotQuantity(item.MinimumQuantity)
	if err != nil {
		return nil, fmt.Errorf("minimum_quantity: %w", err)
	}
	pack, err := parseSnapshotQuantity(item.PackageQuantity)
	if err != nil {
		return nil, fmt.Errorf("package_quantity: %w", err)
	}
	var imageURL *url.URL
	if item.ExternalImageURL != nil {
		imageURL, err = url.Parse(*item.ExternalImageURL)
		if err != nil {
			return nil, fmt.Errorf("external_image_url: %w", err)
		}
	}
	return &domain.ItemStock{
		Item: domain.InventoryItem{
			ID:               item.ID,
			SpaceID:          item.SpaceID,
			CatalogProductID: item.CatalogProductID,
			GTIN:             item.GTIN,
			Name:             item.Name,
			Brand:            item.Brand,
			Category:         item.Category,
			Unit:             item.Unit,
			MinimumQuantity:  minimum,
			PackageQuantity:  pack,
			ExternalImageURL: imageURL,
			ImageObjectKey:   item.ImageObjectKey,
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
			DeletedAt:        item.DeletedAt,
		},
		TotalQuantity: *total,
	}, nil
}

func (batch storedBatch) toDomain() (domain.Batch, error) {
	remaining, err := quantity.ParseQuantity(batch.RemainingQuantity)
	if err != nil {
		return domain.Batch{}, fmt.Errorf("remaining_quantity: %w", err)
	}
	purchasedOn, err := parseSnapshotDate(batch.PurchasedOn)
	if err != nil {
		return domain.Batch{}, fmt.Errorf("purchased_on: %w", err)
	}
	expiresOn, err := parseSnapshotDate(batch.ExpiresOn)
	if err != nil {
		return domain.Batch{}, fmt.Errorf("expires_on: %w", err)
	}
	return domain.Batch{
		ID:                batch.ID,
		SpaceID:           batch.SpaceID,
		ItemID:            batch.ItemID,
		RemainingQuantity: *remaining,
		PurchasedOn:       purchasedOn,
		AddedAt:           batch.AddedAt,
		ExpiresOn:         expiresOn,
	}, nil
}

func snapshotStockOperation(operation *domain.StockOperation) *storedStockOperation {
	if operation == nil {
		return nil
	}
	result := storedStockOperation{
		ID:          operation.ID,
		SpaceID:     operation.SpaceID,
		ItemID:      operation.ItemID,
		ActorUserID: operation.ActorUserID,
		Type:        operation.Type,
		Quantity:    operation.Quantity.String(),
		OccurredAt:  operation.OccurredAt,
		Lines:       make([]storedStockOperationLine, 0, len(operation.Lines)),
	}
	for _, line := range operation.Lines {
		result.Lines = append(result.Lines, storedStockOperationLine{
			SpaceID:     line.SpaceID,
			OperationID: line.OperationID,
			ItemID:      line.ItemID,
			BatchID:     line.BatchID,
			Quantity:    line.Quantity.String(),
		})
	}
	return &result
}

func (operation *storedStockOperation) toDomain() (*domain.StockOperation, error) {
	if operation == nil {
		return nil, nil
	}
	amount, err := quantity.ParseQuantity(operation.Quantity)
	if err != nil {
		return nil, fmt.Errorf("quantity: %w", err)
	}
	lines := make([]domain.StockOperationLine, 0, len(operation.Lines))
	for i, line := range operation.Lines {
		lineAmount, err := quantity.ParseQuantity(line.Quantity)
		if err != nil {
			return nil, fmt.Errorf("lines[%d].quantity: %w", i, err)
		}
		lines = append(lines, domain.StockOperationLine{
			SpaceID:     line.SpaceID,
			OperationID: line.OperationID,
			ItemID:      line.ItemID,
			BatchID:     line.BatchID,
			Quantity:    *lineAmount,
		})
	}
	return &domain.StockOperation{
		ID:          operation.ID,
		SpaceID:     operation.SpaceID,
		ItemID:      operation.ItemID,
		ActorUserID: operation.ActorUserID,
		Type:        operation.Type,
		Quantity:    *amount,
		OccurredAt:  operation.OccurredAt,
		Lines:       lines,
	}, nil
}

func snapshotQuantity(value *quantity.Quantity) *string {
	if value == nil {
		return nil
	}
	text := value.String()
	return &text
}

func parseSnapshotQuantity(text *string) (*quantity.Quantity, error) {
	if text == nil {
		return nil, nil
	}
	return quantity.ParseQuantity(*text)
}

func snapshotDate(value *domain.Date) *string {
	if value == nil {
		return nil
	}
	text := value.String()
	return &text
}

func parseSnapshotDate(text *string) (*domain.Date, error) {
	if text == nil {
		return nil, nil
	}
	return domain.ParseDate(*text)
}
