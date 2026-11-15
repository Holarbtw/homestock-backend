package ports

import (
	"net/url"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

// Optional preserves whether a field was supplied. Set=false means unchanged;
// Set=true with Value=nil means an explicit null, checked by the service.
type Optional[T any] struct {
	Set   bool
	Value *T
}

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

type UpdateItemInput struct {
	ItemID          uuid.UUID
	Name            Optional[string]
	Brand           Optional[string]
	Category        Optional[domain.Category]
	MinimumQuantity Optional[quantity.Quantity]
}

type UpdateItemParams struct {
	Name            string
	Brand           *string
	Category        domain.Category
	MinimumQuantity *quantity.Quantity
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
