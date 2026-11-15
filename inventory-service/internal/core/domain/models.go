package domain

import (
	"net/url"
	"time"

	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

type Category string

const (
	Food      Category = "FOOD"
	Medicine  Category = "MEDICINE"
	Household Category = "HOUSEHOLD"
)

type Unit string

const (
	Milliliter Unit = "ml"
	Gram       Unit = "g"
	Pcs        Unit = "pcs"
)

type StockOperationType string

const (
	Replenish StockOperationType = "REPLENISH"
	Consume   StockOperationType = "CONSUME"
	WriteOff  StockOperationType = "WRITE_OFF"
)

type Space struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
	CreatedAt   time.Time
}

type InventoryItem struct {
	ID               uuid.UUID
	SpaceID          uuid.UUID
	CatalogProductID *uuid.UUID
	GTIN             *string
	Name             string
	Brand            *string
	Category         Category
	Unit             Unit
	MinimumQuantity  *quantity.Quantity
	PackageQuantity  *quantity.Quantity
	ExternalImageURL *url.URL
	ImageObjectKey   *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func (i *InventoryItem) IsLowStock(total quantity.Quantity) bool {
	return i.DeletedAt == nil && i.MinimumQuantity != nil && total.Thousandths() <= i.MinimumQuantity.Thousandths()
}

type Batch struct {
	ID                uuid.UUID
	SpaceID           uuid.UUID
	ItemID            uuid.UUID
	RemainingQuantity quantity.Quantity
	PurchasedOn       *Date
	AddedAt           time.Time
	ExpiresOn         *Date
}

type StockOperation struct {
	ID          uuid.UUID
	SpaceID     uuid.UUID
	ItemID      uuid.UUID
	ActorUserID uuid.UUID
	Type        StockOperationType
	Quantity    quantity.Quantity
	OccurredAt  time.Time
	Lines       []StockOperationLine
}

type StockOperationLine struct {
	SpaceID     uuid.UUID
	OperationID uuid.UUID
	ItemID      uuid.UUID
	BatchID     uuid.UUID
	Quantity    quantity.Quantity
}

type ItemStock struct {
	Item          InventoryItem
	TotalQuantity quantity.Quantity
}

func (s *ItemStock) IsLowStock() bool {
	return s.Item.IsLowStock(s.TotalQuantity)
}
