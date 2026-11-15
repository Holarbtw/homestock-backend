package services

import (
	"context"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

type ItemService struct {
	txManager ports.TxManager
	itemRepo ports.ItemRepository
	spaceRepo ports.SpaceRepository
	batchRepo ports.BatchRepository
}

func NewItemService(
	txManager ports.TxManager,
	itemRepo ports.ItemRepository,
	spaceRepo ports.SpaceRepository,
	batchRepo ports.BatchRepository,
	) *ItemService{
	return &ItemService{
		txManager: txManager,
		itemRepo: itemRepo,
		spaceRepo: spaceRepo,
		batchRepo: batchRepo,
	}
}

func (s *ItemService) UpdateItem(ctx context.Context, userID uuid.UUID, input *ports.UpdateItemInput) (*domain.ItemStock, error){
	var errFields []apperrors.FieldError

	space, err := s.spaceRepo.GetByOwnerUserID(ctx, userID)
	if err != nil {
		return nil, err
	}


	if input.Name.Set{
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
	if input.Brand.Set && input.Brand.Value != nil{
		validBrand, err := validateItemBrand(*input.Brand.Value)

		errFields, err = appendFieldError(errFields, "brand", err)
		if err != nil {
			return nil, err
		}

		input.Brand.Value = &validBrand
	}
	if input.Category.Set{
		if input.Category.Value == nil{
			errFields = append(errFields, apperrors.FieldError{
				Field: "category",
				Code:  domain.ErrInvalidItemCategory.Code,
			})
		}else{
			validCategory, err := validateItemCategory(string(*input.Category.Value))
	
			errFields, err = appendFieldError(errFields, "category", err)
			if err != nil {
				return nil, err
			}
	
			input.Category.Value = &validCategory
		}
	}

	if len(errFields) > 0{
		return nil, domain.ErrInvalidItemUpdate.WithFields(errFields)
	}

	var updatedItem *domain.InventoryItem
	var totalQuantity quantity.Quantity
	txErr := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error{
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
		Item: *updatedItem,
		TotalQuantity: totalQuantity,
	}, nil
}