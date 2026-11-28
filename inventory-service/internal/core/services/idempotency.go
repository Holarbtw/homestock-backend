package services

import (
	"context"
	"fmt"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/google/uuid"
)

func reserveOrLoadResult(
	idempotencyRepo ports.IdempotencyRepository,
	ctx context.Context,
	spaceID uuid.UUID,
	idempotencyKey string,
	fingerprint Fingerprint,
) (*ports.IdempotencyResult, error){
	reserved, err := idempotencyRepo.TryReserve(ctx, spaceID, idempotencyKey, string(fingerprint.Command), fingerprint.Hash)
	if err != nil {
		return nil, err
	}	

	if !reserved{
		record, err := idempotencyRepo.GetByKey(ctx, spaceID, idempotencyKey)
		if err != nil {
			return nil, err
		}

		if record.Result == nil{
			return nil, apperrors.ErrInternalServer.WithInternal(fmt.Errorf(
				"idempotency record has no saved result: space_id=%s request_key=%q command=%s",
				record.SpaceID, record.RequestKey, record.CommandType,
			))
		}

		if record.RequestHash == fingerprint.Hash && record.CommandType == string(fingerprint.Command) {
			return record.Result, nil
		}

		return nil, domain.ErrIdempotencyKeyConflict
	}

	return nil, nil
}