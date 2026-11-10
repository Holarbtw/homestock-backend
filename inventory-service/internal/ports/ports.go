package ports

import (
	"context"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

type TxManager interface {
	WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

type UpdateItemParams struct {
	Name            string
	Brand           *string
	Category        domain.Category
	MinimumQuantity *quantity.Quantity
}

type ItemRepository interface {
	Create(ctx context.Context, item domain.InventoryItem) (*domain.InventoryItem, error)
	GetByID(ctx context.Context, spaceID, itemID uuid.UUID) (*domain.InventoryItem, error)

	GetByIDForUpdate(ctx context.Context, spaceID, itemID uuid.UUID) (*domain.InventoryItem, error)

	ListBySpace(ctx context.Context, spaceID uuid.UUID, limit, offset int) ([]domain.InventoryItem, error)

	Update(ctx context.Context, spaceID, itemID uuid.UUID, params UpdateItemParams) (*domain.InventoryItem, error)

	Delete(ctx context.Context, spaceID, itemID uuid.UUID) error
}
