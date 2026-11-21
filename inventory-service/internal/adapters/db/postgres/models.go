package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type spaceDB struct {
	ID          uuid.UUID `db:"id"`
	OwnerUserID uuid.UUID `db:"owner_user_id"`
	CreatedAt   time.Time `db:"created_at"`
}

type idempotencyRecordDB struct {
	SpaceID     uuid.UUID `db:"space_id"`
	RequestKey  string    `db:"request_key"`
	CommandType string    `db:"command_type"`
	RequestHash []byte    `db:"request_hash"`
	ResultData  []byte    `db:"result_data"`
	CreatedAt   time.Time `db:"created_at"`
}

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

type itemStockDB struct {
	itemDB
	TotalQuantity string `db:"total_quantity"`
}

type batchDB struct {
	ID                uuid.UUID   `db:"id"`
	SpaceID           uuid.UUID   `db:"space_id"`
	ItemID            uuid.UUID   `db:"item_id"`
	RemainingQuantity string      `db:"remaining_quantity"`
	PurchasedOn       pgtype.Date `db:"purchased_on"`
	AddedAt           time.Time   `db:"added_at"`
	ExpiresOn         pgtype.Date `db:"expires_on"`
}

type stockOperationDB struct {
	ID          uuid.UUID `db:"id"`
	SpaceID     uuid.UUID `db:"space_id"`
	ItemID      uuid.UUID `db:"item_id"`
	ActorUserID uuid.UUID `db:"actor_user_id"`
	Type        string    `db:"type"`
	Quantity    string    `db:"quantity"`
	OccurredAt  time.Time `db:"occurred_at"`
}
