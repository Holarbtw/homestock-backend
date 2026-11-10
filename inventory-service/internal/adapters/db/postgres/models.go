package postgres

import (
	"time"

	"github.com/google/uuid"
)

type itemDB struct {
	ID               uuid.UUID  `db:"id"`
	SpaceID          uuid.UUID  `db:"space_id"`
	CatalogProductID *uuid.UUID `db:"catalog_product_id"`
	GTIN             *string    `db:"gtin"`
	Name             string     `db:"name"`
	Brand            *string    `db:"brand"`
	Category         string     `db:"category"`
	Unit             string     `db:"unit"`
	MinimumQuantity  *string    `db:"minimum_quantity"`
	PackageQuantity  *string    `db:"package_quantity"`
	ExternalImageURL *string    `db:"external_image_url"`
	ImageObjectKey   *string    `db:"image_object_key"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	DeletedAt        *time.Time `db:"deleted_at"`
}
