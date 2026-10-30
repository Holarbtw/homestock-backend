package postgres

import (
	"time"

	"github.com/google/uuid"
)

type productDB struct {
	ID uuid.UUID                `db:"id"`           
	GTIN string                 `db:"gtin"`        
	Name string                 `db:"name"`
	Brand *string               `db:"brand"`     
	Category *string            `db:"category"`     
	PackageQuantity *string     `db:"package_quantity"`     
	PackageUnit *string         `db:"package_unit"`   
	ImageURL *string         `db:"image_url"`   
	CreatedAt time.Time         `db:"created_at"`   
	UpdatedAt time.Time `db:"updated_at"`
}