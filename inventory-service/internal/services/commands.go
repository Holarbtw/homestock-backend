package services

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
)

type CommandType string

const (
	CommandCreateItem    CommandType = "CREATE_ITEM"
	CommandAddBatch      CommandType = "ADD_BATCH"
	CommandDecreaseItem  CommandType = "DECREASE_ITEM"
	CommandDecreaseBatch CommandType = "DECREASE_BATCH"
	CommandUpdateItem    CommandType = "UPDATE_ITEM"
)

func serializeCommand(value any) (*fingerprintEnvelope, error) {
	var data fingerprintEnvelope
	switch input := value.(type) {
	case *CreateItemInput:
		if input == nil {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("nil value passed in serializeCommand"),
			)
		}
		fingerprint := createItemFingerprint{
			Name:     input.Name,
			Brand:    input.Brand,
			GTIN:     input.GTIN,
			Category: input.Category,
			Unit:     input.Unit,
		}

		if input.CatalogProductID != nil {
			catalogProductID := input.CatalogProductID.String()
			fingerprint.CatalogProductID = &catalogProductID
		}
		if input.MinimumQuantity != nil {
			minQuantity := input.MinimumQuantity.String()
			fingerprint.MinimumQuantity = &minQuantity
		}
		if input.PackageQuantity != nil {
			packageQuantity := input.PackageQuantity.String()
			fingerprint.PackageQuantity = &packageQuantity
		}
		if input.ExternalImageURL != nil {
			externalImageURL := input.ExternalImageURL.String()
			fingerprint.ExternalImageURL = &externalImageURL
		}
		if input.FirstBatch != nil {
			quantity := input.FirstBatch.Quantity.String()
			firstBatch := firstBatchFingerprint{
				Quantity: quantity,
			}
			if input.FirstBatch.PurchasedOn != nil {
				purchasedOn := input.FirstBatch.PurchasedOn.String()
				firstBatch.PurchasedOn = &purchasedOn
			}
			if input.FirstBatch.ExpiresOn != nil {
				expiresOn := input.FirstBatch.ExpiresOn.String()
				firstBatch.ExpiresOn = &expiresOn
			}
			fingerprint.FirstBatch = &firstBatch
		}

		bytes, err := json.Marshal(fingerprint)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}

		data = fingerprintEnvelope{Bytes: bytes, Command: CommandCreateItem}
	case *AddBatchInput:
		if input == nil {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("nil value passed in serializeCommand"),
			)
		}
		fingerprint := addBatchFingerprint{
			ItemID:   input.ItemID.String(),
			Quantity: input.Quantity.String(),
		}
		if input.PurchasedOn != nil {
			purchasedOn := input.PurchasedOn.String()
			fingerprint.PurchasedOn = &purchasedOn
		}
		if input.ExpiresOn != nil {
			expiresOn := input.ExpiresOn.String()
			fingerprint.ExpiresOn = &expiresOn
		}

		bytes, err := json.Marshal(fingerprint)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}

		data = fingerprintEnvelope{Bytes: bytes, Command: CommandAddBatch}
	case *DecreaseItemInput:
		if input == nil {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("nil value passed in serializeCommand"),
			)
		}
		fingerprint := decreaseItemFingerprint{
			ItemID:   input.ItemID.String(),
			Type:     input.Type,
			Quantity: input.Quantity.String(),
		}

		bytes, err := json.Marshal(fingerprint)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}

		data = fingerprintEnvelope{Bytes: bytes, Command: CommandDecreaseItem}
	case *DecreaseBatchInput:
		if input == nil {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("nil value passed in serializeCommand"),
			)
		}
		fingerprint := decreaseBatchFingerprint{
			ItemID:   input.ItemID.String(),
			BatchID:  input.BatchID.String(),
			Type:     input.Type,
			Quantity: input.Quantity.String(),
		}

		bytes, err := json.Marshal(fingerprint)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}

		data = fingerprintEnvelope{Bytes: bytes, Command: CommandDecreaseBatch}
	default:
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("unsupported input object type: %T", value),
		)
	}

	return &data, nil
}

func hashFingerprint(data []byte) [sha256.Size]byte{
	return sha256.Sum256(data)
}