package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

var _ ports.ItemService = (*ItemService)(nil)

type ItemService struct {
	txManager       ports.TxManager
	itemRepo        ports.ItemRepository
	spaceRepo       ports.SpaceRepository
	batchRepo       ports.BatchRepository
	idempotencyRepo ports.IdempotencyRepository
	stockRepo       ports.StockOperationRepository
}

func NewItemService(
	txManager ports.TxManager,
	itemRepo ports.ItemRepository,
	spaceRepo ports.SpaceRepository,
	batchRepo ports.BatchRepository,
	idempotencyRepo ports.IdempotencyRepository,
	stockRepo ports.StockOperationRepository,
) *ItemService {
	return &ItemService{
		txManager:       txManager,
		itemRepo:        itemRepo,
		spaceRepo:       spaceRepo,
		batchRepo:       batchRepo,
		idempotencyRepo: idempotencyRepo,
		stockRepo:       stockRepo,
	}
}

func (s *ItemService) UpdateItem(ctx context.Context, userID uuid.UUID, input *ports.UpdateItemInput) (*domain.ItemStock, error) {
	var errFields []apperrors.FieldError

	space, err := s.spaceRepo.GetByOwnerUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Name.Set {
		if input.Name.Value == nil {
			errFields = append(errFields, apperrors.FieldError{
				Field: "name",
				Code:  domain.ErrInvalidItemName.Code,
			})
		} else {
			validName, err := validateItemName(*input.Name.Value)

			errFields, err = appendFieldError(errFields, "name", err)
			if err != nil {
				return nil, err
			}

			input.Name.Value = &validName
		}
	}
	if input.Brand.Set && input.Brand.Value != nil {
		validBrand, err := validateItemBrand(*input.Brand.Value)

		errFields, err = appendFieldError(errFields, "brand", err)
		if err != nil {
			return nil, err
		}

		input.Brand.Value = &validBrand
	}
	if input.Category.Set {
		if input.Category.Value == nil {
			errFields = append(errFields, apperrors.FieldError{
				Field: "category",
				Code:  domain.ErrInvalidItemCategory.Code,
			})
		} else {
			validCategory, err := validateItemCategory(string(*input.Category.Value))

			errFields, err = appendFieldError(errFields, "category", err)
			if err != nil {
				return nil, err
			}

			input.Category.Value = &validCategory
		}
	}

	if len(errFields) > 0 {
		return nil, domain.ErrInvalidItemUpdate.WithFields(errFields)
	}

	var updatedItem *domain.InventoryItem
	var totalQuantity quantity.Quantity
	txErr := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		item, err := s.itemRepo.GetByIDForUpdate(txCtx, space.ID, input.ItemID)
		if err != nil {
			return err
		}

		params := ports.UpdateItemParams{
			Name:            item.Name,
			Brand:           item.Brand,
			Category:        item.Category,
			MinimumQuantity: item.MinimumQuantity,
		}

		if input.Name.Set {
			params.Name = *input.Name.Value
		}
		if input.Brand.Set {
			params.Brand = input.Brand.Value
		}
		if input.Category.Set {
			params.Category = *input.Category.Value
		}
		if input.MinimumQuantity.Set {
			params.MinimumQuantity = input.MinimumQuantity.Value
		}

		updatedItem, err = s.itemRepo.Update(txCtx, space.ID, input.ItemID, params)
		if err != nil {
			return err
		}

		totalQuantity, err = s.batchRepo.GetTotalQuantity(txCtx, space.ID, input.ItemID)
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return &domain.ItemStock{
		Item:          *updatedItem,
		TotalQuantity: totalQuantity,
	}, nil
}

func (s *ItemService) CreateItem(
	ctx context.Context,
	userID uuid.UUID,
	idempotencyKey string,
	input *ports.CreateItemInput,
) (stock *domain.ItemStock, batches []domain.Batch, operation *domain.StockOperation, err error) {
	normalizedInput, err := validateCreateItemInput(input, idempotencyKey, time.Now())
	if err != nil {
		return nil, nil, nil, err
	}

	space, err := s.spaceRepo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, nil, nil, err
	}

	txErr := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		fingerprint, err := buildFingerprint(normalizedInput)
		if err != nil {
			return err
		}

		reserved, err := s.idempotencyRepo.TryReserve(txCtx, space.ID, idempotencyKey, string(fingerprint.Command), fingerprint.Hash)
		if err != nil {
			return err
		}

		if !reserved {
			record, err := s.idempotencyRepo.GetByKey(txCtx, space.ID, idempotencyKey)
			if err != nil {
				return err
			}

			if record.Result == nil {
				return apperrors.ErrInternalServer
			}

			if record.RequestHash == fingerprint.Hash && record.CommandType == string(fingerprint.Command) {
				decodedItem, decodedBatches, decodedStockOperation, err := decodeCreateItemResult(record.Result.Data)
				if err != nil {
					return err
				}
				stock = decodedItem
				batches = decodedBatches
				operation = decodedStockOperation

				return nil
			}

			return domain.ErrIdempotencyKeyConflict
		}

		item, err := s.itemRepo.Create(txCtx, domain.InventoryItem{
			SpaceID:          space.ID,
			CatalogProductID: normalizedInput.CatalogProductID,
			GTIN:             normalizedInput.GTIN,
			Name:             normalizedInput.Name,
			Brand:            normalizedInput.Brand,
			Category:         normalizedInput.Category,
			Unit:             normalizedInput.Unit,
			MinimumQuantity:  normalizedInput.MinimumQuantity,
			PackageQuantity:  normalizedInput.PackageQuantity,
			ExternalImageURL: normalizedInput.ExternalImageURL,
		})
		if err != nil {
			return err
		}

		stock = &domain.ItemStock{Item: *item}
		batches = make([]domain.Batch, 0, 1)

		if normalizedInput.FirstBatch != nil {
			batch, err := s.batchRepo.Create(txCtx, domain.Batch{
				SpaceID:           space.ID,
				ItemID:            item.ID,
				RemainingQuantity: normalizedInput.FirstBatch.Quantity,
				PurchasedOn:       normalizedInput.FirstBatch.PurchasedOn,
				ExpiresOn:         normalizedInput.FirstBatch.ExpiresOn,
			})

			if err != nil {
				return err
			}

			op, err := s.stockRepo.CreateReplenishment(txCtx, batch, userID)
			if err != nil {
				return err
			}

			batches = append(batches, *batch)
			operation = op
			stock.TotalQuantity = batch.RemainingQuantity
		}

		data, err := encodeCreateItemResult(stock, batches, operation)
		if err != nil {
			return err
		}

		return s.idempotencyRepo.SaveResult(
			txCtx, space.ID, idempotencyKey,
			ports.IdempotencyResult{Data: data},
		)
	})

	if txErr != nil {
		return nil, nil, nil, txErr
	}

	return stock, batches, operation, nil
}

func (s *ItemService) GetItem(ctx context.Context, userID, itemID uuid.UUID) (*domain.ItemStock, []domain.Batch, error) {
	space, err := s.spaceRepo.GetByOwnerUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	item, err := s.itemRepo.GetByID(ctx, space.ID, itemID)
	if err != nil {
		return nil, nil, err
	}
	batches, err := s.batchRepo.ListByItem(ctx, space.ID, itemID)
	if err != nil {
		return nil, nil, err
	}

	var totalThousandths int64
	for _, batch := range batches {
		remaining := batch.RemainingQuantity.Thousandths()
		if remaining > math.MaxInt64-totalThousandths {
			return nil, nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("total batch quantity for item %s exceeds the supported range", itemID),
			)
		}
		totalThousandths += remaining
	}
	total, err := quantity.NewQuantity(totalThousandths)
	if err != nil {
		return nil, nil, apperrors.ErrInternalServer.WithInternal(err)
	}
	return &domain.ItemStock{Item: *item, TotalQuantity: *total}, batches, nil
}

func (s *ItemService) ListItems(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.ItemStock, error) {
	if limit <= 0 || offset < 0 {
		return nil, apperrors.ErrInvalidRequest
	}
	space, err := s.spaceRepo.GetByOwnerUserID(ctx, userID)
	if errors.Is(err, domain.ErrSpaceNotFound) {
		return []domain.ItemStock{}, nil
	}
	if err != nil {
		return nil, err
	}
	return s.itemRepo.ListBySpace(ctx, space.ID, limit, offset)
}

func (s *ItemService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	space, err := s.spaceRepo.GetByOwnerUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.itemRepo.Delete(ctx, space.ID, itemID)
}
