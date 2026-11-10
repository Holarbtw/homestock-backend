package services

import (
	"crypto/sha256"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/domain"
)

type fingerprintEnvelope struct {
	Bytes   []byte
	Command CommandType
}

type createItemFingerprint struct {
	Name             string
	Brand            *string
	CatalogProductID *string
	GTIN             *string
	Category         domain.Category
	Unit             domain.Unit
	MinimumQuantity  *string
	PackageQuantity  *string
	ExternalImageURL *string
	FirstBatch       *firstBatchFingerprint
}

type firstBatchFingerprint struct {
	Quantity    string
	PurchasedOn *string
	ExpiresOn   *string
}

type addBatchFingerprint struct {
	ItemID      string
	Quantity    string
	PurchasedOn *string
	ExpiresOn   *string
}

type decreaseItemFingerprint struct {
	ItemID   string
	Type     domain.StockOperationType
	Quantity string
}

type decreaseBatchFingerprint struct {
	ItemID   string
	BatchID  string
	Type     domain.StockOperationType
	Quantity string
}

type Fingerprint struct {
	Command CommandType
	Hash    [sha256.Size]byte
}

func buildFingerprint(input any) (Fingerprint, error){
	envelop, err := serializeCommand(input)
	if err != nil {
		return Fingerprint{}, err
	}

	hash := hashFingerprint(envelop.Bytes)
	return Fingerprint{
		Command: envelop.Command,
		Hash: hash,
	}, nil
}