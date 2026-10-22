package domain

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

type Product struct {
	ID uuid.UUID
	GTIN string
	Name string
	Brand *string
	Category *string
	PackageSize *PackageSize
	ImageURL *url.URL
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PackageSize struct{
	Quantity quantity.Quantity
	Unit Unit
}

type Unit string

const(
	Milliliter Unit = "ml"
	Gram Unit = "g"
	Pcs Unit = "pcs"
)

type ProductSource struct{
	ID uuid.UUID
	ProductID uuid.UUID
	Provider string
	ExternalCode string
	RawPayload json.RawMessage
	FetchedAt time.Time
	SourceUpdatedAt *time.Time
}