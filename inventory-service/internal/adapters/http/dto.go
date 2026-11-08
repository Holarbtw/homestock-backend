package http

import (
	"time"

	"github.com/google/uuid"
)

type CreateItemRequest struct {
	Name             string           `json:"name" validate:"required" example:"Молоко 930 мл"`
	Brand            *string          `json:"brand,omitempty" example:"Простоквашино"`
	CatalogProductID *uuid.UUID       `json:"catalog_product_id,omitempty" format:"uuid"`
	GTIN             *string          `json:"gtin,omitempty" example:"04600634855338"`
	Category         string           `json:"category" validate:"required" enums:"FOOD,MEDICINE,HOUSEHOLD" example:"FOOD"`
	Unit             string           `json:"unit" validate:"required" enums:"ml,g,pcs" example:"ml"`
	MinimumQuantity  *string          `json:"minimum_quantity,omitempty" example:"500.000"`
	PackageQuantity  *string          `json:"package_quantity,omitempty" example:"930.000"`
	ExternalImageURL *string          `json:"external_image_url,omitempty" format:"uri"`
	FirstBatch       *AddBatchRequest `json:"first_batch,omitempty"`
}

type AddBatchRequest struct {
	Quantity    string  `json:"quantity" validate:"required" example:"1860.000"`
	PurchasedOn *string `json:"purchased_on,omitempty" format:"date" example:"2026-11-07"`
	ExpiresOn   *string `json:"expires_on,omitempty" format:"date" example:"2026-11-15"`
}

type ConsumeRequest struct {
	Quantity string `json:"quantity" validate:"required" example:"250.000"`
}

type WriteOffRequest struct {
	Quantity string `json:"quantity" validate:"required" example:"100.000"`
}

type ItemResponse struct {
	ID               uuid.UUID  `json:"id" format:"uuid" validate:"required"`
	SpaceID          uuid.UUID  `json:"space_id" format:"uuid" validate:"required"`
	CatalogProductID *uuid.UUID `json:"catalog_product_id,omitempty" format:"uuid"`
	GTIN             *string    `json:"gtin,omitempty" example:"04600634855338"`
	Name             string     `json:"name" validate:"required" example:"Молоко 930 мл"`
	Brand            *string    `json:"brand,omitempty" example:"Простоквашино"`
	Category         string     `json:"category" validate:"required" enums:"FOOD,MEDICINE,HOUSEHOLD" example:"FOOD"`
	Unit             string     `json:"unit" validate:"required" enums:"ml,g,pcs" example:"ml"`
	MinimumQuantity  *string    `json:"minimum_quantity,omitempty" example:"500.000"`
	PackageQuantity  *string    `json:"package_quantity,omitempty" example:"930.000"`
	ImageURL         *string    `json:"image_url,omitempty" format:"uri"`
	TotalQuantity    string     `json:"total_quantity" validate:"required" example:"1860.000"`
	IsLowStock       bool       `json:"is_low_stock" example:"false"`
	CreatedAt        time.Time  `json:"created_at" format:"date-time" validate:"required"`
	UpdatedAt        time.Time  `json:"updated_at" format:"date-time" validate:"required"`
}

type BatchResponse struct {
	ID                uuid.UUID `json:"id" format:"uuid" validate:"required"`
	SpaceID           uuid.UUID `json:"space_id" format:"uuid" validate:"required"`
	ItemID            uuid.UUID `json:"item_id" format:"uuid" validate:"required"`
	RemainingQuantity string    `json:"remaining_quantity" validate:"required" example:"1860.000"`
	PurchasedOn       *string   `json:"purchased_on,omitempty" format:"date" example:"2026-11-07"`
	ExpiresOn         *string   `json:"expires_on,omitempty" format:"date" example:"2026-11-15"`
	AddedAt           time.Time `json:"added_at" format:"date-time" validate:"required"`
}

type StockOperationResponse struct {
	ID          uuid.UUID                    `json:"id" format:"uuid" validate:"required"`
	SpaceID     uuid.UUID                    `json:"space_id" format:"uuid" validate:"required"`
	ItemID      uuid.UUID                    `json:"item_id" format:"uuid" validate:"required"`
	ActorUserID uuid.UUID                    `json:"actor_user_id" format:"uuid" validate:"required"`
	Type        string                       `json:"type" validate:"required" enums:"REPLENISH,CONSUME,WRITE_OFF" example:"CONSUME"`
	Quantity    string                       `json:"quantity" validate:"required" example:"250.000"`
	OccurredAt  time.Time                    `json:"occurred_at" format:"date-time" validate:"required"`
	Lines       []StockOperationLineResponse `json:"lines" validate:"required"`
}

type StockOperationLineResponse struct {
	BatchID  uuid.UUID `json:"batch_id" format:"uuid" validate:"required"`
	Quantity string    `json:"quantity" validate:"required" example:"250.000"`
}

type ItemDetailsResponse struct {
	Item ItemResponse `json:"item" validate:"required"`
	Batches []BatchResponse `json:"batches" validate:"required"`
}

type StockChangeResponse struct {
	Item ItemResponse `json:"item" validate:"required"`
	Batches   []BatchResponse        `json:"batches" validate:"required"`
	Operation StockOperationResponse `json:"operation" validate:"required"`
}

type CreateItemResponse struct {
	Item      ItemResponse            `json:"item" validate:"required"`
	Batches   []BatchResponse         `json:"batches" validate:"required"`
	Operation *StockOperationResponse `json:"operation,omitempty"`
}
