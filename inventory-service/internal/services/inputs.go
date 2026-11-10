package services

import (
	"net/url"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

type CreateItemInput struct {
	Name             string
	Brand            *string
	CatalogProductID *uuid.UUID
	GTIN             *string
	Category         domain.Category
	Unit             domain.Unit
	MinimumQuantity  *quantity.Quantity
	PackageQuantity  *quantity.Quantity
	ExternalImageURL *url.URL
	FirstBatch       *InitialBatchInput
}

type InitialBatchInput struct {
	Quantity    quantity.Quantity
	PurchasedOn *domain.Date
	ExpiresOn   *domain.Date
}

type AddBatchInput struct {
	ItemID      uuid.UUID
	Quantity    quantity.Quantity
	PurchasedOn *domain.Date
	ExpiresOn   *domain.Date
}

type DecreaseItemInput struct {
	ItemID   uuid.UUID
	Type     domain.StockOperationType
	Quantity quantity.Quantity
}

type DecreaseBatchInput struct {
	ItemID   uuid.UUID
	BatchID  uuid.UUID
	Type     domain.StockOperationType
	Quantity quantity.Quantity
}
