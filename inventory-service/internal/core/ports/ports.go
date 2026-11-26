package ports

import (
	"context"
	"crypto/sha256"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

type TxManager interface {
	WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

type SpaceRepository interface {
	GetByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) (*domain.Space, error)
	GetOrCreate(ctx context.Context, ownerUserID uuid.UUID) (*domain.Space, error)
}

type ItemRepository interface {
	Create(ctx context.Context, item domain.InventoryItem) (*domain.InventoryItem, error)
	GetByID(ctx context.Context, spaceID, itemID uuid.UUID) (*domain.InventoryItem, error)

	GetByIDForUpdate(ctx context.Context, spaceID, itemID uuid.UUID) (*domain.InventoryItem, error)

	ListBySpace(ctx context.Context, spaceID uuid.UUID, limit, offset int) ([]domain.ItemStock, error)

	Update(ctx context.Context, spaceID, itemID uuid.UUID, params UpdateItemParams) (*domain.InventoryItem, error)

	Delete(ctx context.Context, spaceID, itemID uuid.UUID) error
}

type BatchRepository interface {
	// Create returns the batch with its database-generated ID and addition time.
	Create(ctx context.Context, batch domain.Batch) (*domain.Batch, error)
	ListByItem(ctx context.Context, spaceID, itemID uuid.UUID) ([]domain.Batch, error)
	// GetTotalQuantity includes expired batches; the service checks item access and active state.
	GetTotalQuantity(ctx context.Context, spaceID, itemID uuid.UUID) (quantity.Quantity, error)
}

type StockOperationRepository interface {
	// CreateReplenishment records a newly created batch and its initial quantity in the caller's transaction.
	CreateReplenishment(ctx context.Context, batch *domain.Batch, actorUserID uuid.UUID) (*domain.StockOperation, error)
}

type IdempotencyRepository interface {
	// TryReserve requires a transaction. False means the key already exists, regardless of its fingerprint.
	// A new reservation must be finalized with SaveResult before that transaction commits.
	TryReserve(ctx context.Context, spaceID uuid.UUID, requestKey, commandType string, requestHash [sha256.Size]byte) (bool, error)
	GetByKey(ctx context.Context, spaceID uuid.UUID, requestKey string) (*IdempotencyRecord, error)
	// SaveResult finalizes a reservation once, in the same transaction as the business changes.
	SaveResult(ctx context.Context, spaceID uuid.UUID, requestKey string, result IdempotencyResult) error
}

type ItemService interface {
	CreateItem(
		ctx context.Context,
		userID uuid.UUID,
		idempotencyKey string,
		input *CreateItemInput,
	) (stock *domain.ItemStock, batches []domain.Batch, operation *domain.StockOperation, err error)

	GetItem(ctx context.Context, userID, itemID uuid.UUID) (stock *domain.ItemStock, batches []domain.Batch, err error)
	ListItems(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.ItemStock, error)
	UpdateItem(ctx context.Context, userID uuid.UUID, input *UpdateItemInput) (*domain.ItemStock, error)
	DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error
}

type CatalogLookup interface{
	FindByGTIN(ctx context.Context, gtin string) (*domain.CatalogProduct, error)
}

type CatalogSearchService interface{
	FindByGTIN(ctx context.Context, gtin string) (*domain.CatalogProduct, error)
}