package services

import (
	"context"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/google/uuid"
)

type StockService struct {
	txManager       ports.TxManager
	itemRepo        ports.ItemRepository
	spaceRepo       ports.SpaceRepository
	batchRepo       ports.BatchRepository
	idempotencyRepo ports.IdempotencyRepository
	stockRepo       ports.StockOperationRepository
}

func NewStockService(
	txManager ports.TxManager,
	itemRepo ports.ItemRepository,
	spaceRepo ports.SpaceRepository,
	batchRepo ports.BatchRepository,
idempotencyRepo ports.IdempotencyRepository,
	stockRepo ports.StockOperationRepository,
) *StockService {
	return &StockService{
		txManager:       txManager,
		itemRepo:        itemRepo,
		spaceRepo:       spaceRepo,
		batchRepo:       batchRepo,
		idempotencyRepo: idempotencyRepo,
		stockRepo:       stockRepo,
	}
}

func (s *StockService) Replenish(
	ctx context.Context,
	idempotencyKey string, 
	input *ports.AddBatchInput, 
	userID uuid.UUID,
) (stock *domain.ItemStock, batches []domain.Batch, operation *domain.StockOperation, err error){
	err = validateAddBatchInput(input)
	if err != nil {
		return nil, nil, nil, err
	}
	
	err = validateIdempotencyKey(idempotencyKey)
	if err != nil {
		return nil, nil, nil, err
	}

	space, err := s.spaceRepo.GetByOwnerUserID(ctx, userID)
	if err != nil {
		return nil, nil, nil, err
	}

	txErr := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error{
		fingerprint, err := buildFingerprint(input)
		if err != nil {
			return err
		}

		result, err := reserveOrLoadResult(s.idempotencyRepo, txCtx, space.ID, idempotencyKey, fingerprint)
		if err != nil {
			return err
		}

		if result != nil{
			decodedItem, decodedBatches, decodedStockOperation, err := decodeItemResult(result.Data)
			if err != nil {
				return err
			}

			stock = decodedItem
			batches = decodedBatches
			operation = decodedStockOperation

			return nil
		}

		item, err := s.itemRepo.GetByIDForUpdate(txCtx, space.ID, input.ItemID)
		if err != nil {
			return err
		}

		stock = &domain.ItemStock{Item: *item}

		batches = make([]domain.Batch, 0, 1)

		batch, err := s.batchRepo.Create(txCtx, domain.Batch{
			SpaceID:           space.ID,
			ItemID:            input.ItemID,
			RemainingQuantity: input.Quantity,
			PurchasedOn:       input.PurchasedOn,
			ExpiresOn:         input.ExpiresOn,
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
		total, err := s.batchRepo.GetTotalQuantity(txCtx, space.ID, input.ItemID)
		if err != nil {
			return err
		}

		stock.TotalQuantity = total

		data, err := encodeItemResult(stock, batches, operation)
		if err != nil {
			return err
		}

		return s.idempotencyRepo.SaveResult(
			txCtx, space.ID, idempotencyKey, ports.IdempotencyResult{Data: data})
	})

	if txErr != nil {
		return nil, nil, nil, txErr
	}

	return stock, batches, operation, nil
}