package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"testing"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
)

func TestSerializeCreateItem(t *testing.T) {
	brand, gtin := "Бренд", "04600634855338"
	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	packageQuantity := inputQuantity(t, "930")
	input := &CreateItemInput{
		Name: "Молоко", Brand: &brand, GTIN: &gtin, CatalogProductID: &catalogID,
		Category: domain.Food, Unit: domain.Milliliter, PackageQuantity: &packageQuantity,
		ExternalImageURL: &url.URL{Scheme: "https", Host: "example.test", Path: "/milk.jpg"},
		FirstBatch: &InitialBatchInput{
			Quantity:    inputQuantity(t, "1860"),
			PurchasedOn: inputDate(t, "2026-11-07"), ExpiresOn: inputDate(t, "2026-11-15"),
		},
	}
	envelope := serializedInput(t, input, CommandCreateItem)
	var got createItemFingerprint
	decodeFingerprint(t, envelope.Bytes, &got)
	if got.Name != input.Name || got.Category != domain.Food || got.Unit != domain.Milliliter {
		t.Fatalf("incorrect card fields: %+v", got)
	}
	if got.Brand == nil || *got.Brand != brand || got.GTIN == nil || *got.GTIN != gtin {
		t.Fatalf("brand or GTIN lost: %+v", got)
	}
	if got.CatalogProductID == nil || *got.CatalogProductID != catalogID.String() {
		t.Fatalf("catalog ID lost: %+v", got)
	}
	if got.MinimumQuantity != nil || got.PackageQuantity == nil || *got.PackageQuantity != "930.000" {
		t.Fatalf("package quantity must be independent of the absent threshold: %+v", got)
	}
	if got.ExternalImageURL == nil || *got.ExternalImageURL != "https://example.test/milk.jpg" {
		t.Fatalf("image URL lost: %+v", got)
	}
	if got.FirstBatch == nil || got.FirstBatch.Quantity != "1860.000" {
		t.Fatalf("first batch lost: %+v", got)
	}
	checkFingerprintDate(t, got.FirstBatch.PurchasedOn, input.FirstBatch.PurchasedOn)
	checkFingerprintDate(t, got.FirstBatch.ExpiresOn, input.FirstBatch.ExpiresOn)

	zero := inputQuantity(t, "0")
	input.MinimumQuantity = &zero
	withZero := serializedInput(t, input, CommandCreateItem)
	decodeFingerprint(t, withZero.Bytes, &got)
	if got.MinimumQuantity == nil || *got.MinimumQuantity != "0.000" || bytes.Equal(envelope.Bytes, withZero.Bytes) {
		t.Fatal("explicit zero threshold must differ from an absent threshold")
	}
}

func TestSerializeAddBatch(t *testing.T) {
	itemID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	for _, withDates := range []bool{false, true} {
		input := &AddBatchInput{ItemID: itemID, Quantity: inputQuantity(t, "1.5")}
		if withDates {
			input.PurchasedOn, input.ExpiresOn = inputDate(t, "2026-11-07"), inputDate(t, "2026-11-15")
		}
		var got addBatchFingerprint
		decodeFingerprint(t, serializedInput(t, input, CommandAddBatch).Bytes, &got)
		if got.ItemID != itemID.String() || got.Quantity != "1.500" {
			t.Fatalf("incorrect batch payload: %+v", got)
		}
		checkFingerprintDate(t, got.PurchasedOn, input.PurchasedOn)
		checkFingerprintDate(t, got.ExpiresOn, input.ExpiresOn)
	}
}

func TestSerializeAddBatchReflectsChangedInput(t *testing.T) {
	base := AddBatchInput{
		ItemID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Quantity: inputQuantity(t, "1"),
		PurchasedOn: inputDate(t, "2026-11-07"), ExpiresOn: inputDate(t, "2026-11-15"),
	}
	original := serializedInput(t, &base, CommandAddBatch).Bytes
	for _, tt := range []struct {
		name   string
		change func(*AddBatchInput)
	}{
		{"item", func(v *AddBatchInput) { v.ItemID = uuid.MustParse("00000000-0000-0000-0000-000000000003") }},
		{"quantity", func(v *AddBatchInput) { v.Quantity = inputQuantity(t, "2") }},
		{"purchase date", func(v *AddBatchInput) { v.PurchasedOn = inputDate(t, "2026-11-08") }},
		{"expiry date", func(v *AddBatchInput) { v.ExpiresOn = inputDate(t, "2026-11-16") }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			changed := base
			tt.change(&changed)
			if bytes.Equal(original, serializedInput(t, &changed, CommandAddBatch).Bytes) {
				t.Fatal("changed input produced identical fingerprint bytes")
			}
		})
	}
	equivalent := base
	equivalent.Quantity = inputQuantity(t, "1.000")
	if !bytes.Equal(original, serializedInput(t, &equivalent, CommandAddBatch).Bytes) {
		t.Fatal("equivalent normalized quantities produced different fingerprint bytes")
	}
}

func TestSerializeDecreaseInputs(t *testing.T) {
	itemID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	batchID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	otherID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	amount := inputQuantity(t, "1")
	for _, tt := range []struct {
		name    string
		command CommandType
		build   func(uuid.UUID, uuid.UUID, domain.StockOperationType, quantity.Quantity) any
	}{
		{"item", CommandDecreaseItem, func(itemID, batchID uuid.UUID, reason domain.StockOperationType, amount quantity.Quantity) any {
			return &DecreaseItemInput{ItemID: itemID, Type: reason, Quantity: amount}
		}},
		{"batch", CommandDecreaseBatch, func(itemID, batchID uuid.UUID, reason domain.StockOperationType, amount quantity.Quantity) any {
			return &DecreaseBatchInput{ItemID: itemID, BatchID: batchID, Type: reason, Quantity: amount}
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			original := serializedInput(t, tt.build(itemID, batchID, domain.Consume, amount), tt.command).Bytes
			variants := []struct {
				name            string
				itemID, batchID uuid.UUID
				reason          domain.StockOperationType
				amount          quantity.Quantity
				same            bool
			}{
				{"item", otherID, batchID, domain.Consume, amount, false},
				{"quantity", itemID, batchID, domain.Consume, inputQuantity(t, "2.5"), false},
				{"reason", itemID, batchID, domain.WriteOff, amount, false},
				{"equivalent quantity", itemID, batchID, domain.Consume, inputQuantity(t, "1.000"), true},
			}
			if tt.command == CommandDecreaseBatch {
				variant := variants[0]
				variant.name, variant.itemID, variant.batchID = "batch", itemID, otherID
				variants = append(variants, variant)
			}
			for _, variant := range variants {
				t.Run(variant.name, func(t *testing.T) {
					data := serializedInput(t, tt.build(variant.itemID, variant.batchID, variant.reason, variant.amount), tt.command).Bytes
					var got struct {
						ItemID, BatchID, Quantity string
						Type                      domain.StockOperationType
					}
					decodeFingerprint(t, data, &got)
					wantBatchID := ""
					if tt.command == CommandDecreaseBatch {
						wantBatchID = variant.batchID.String()
					}
					if got.ItemID != variant.itemID.String() || got.BatchID != wantBatchID || got.Quantity != variant.amount.String() || got.Type != variant.reason {
						t.Fatalf("incorrect decrease payload: %+v", got)
					}
					if bytes.Equal(original, data) != variant.same {
						t.Fatalf("fingerprint equality = %v, want %v", bytes.Equal(original, data), variant.same)
					}
				})
			}
		})
	}
}

func TestSerializeInvalidInput(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input any
	}{
		{"nil create", (*CreateItemInput)(nil)},
		{"nil add batch", (*AddBatchInput)(nil)},
		{"nil decrease item", (*DecreaseItemInput)(nil)},
		{"nil decrease batch", (*DecreaseBatchInput)(nil)},
		{"nil interface", nil},
		{"unsupported", "not an input"},
		{"standalone initial batch", &InitialBatchInput{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			envelope, err := serializeCommand(tt.input)
			if envelope != nil || err == nil {
				t.Fatalf("expected no envelope and an internal error, got %+v, %v", envelope, err)
			}
			var appErr *apperrors.CustomError
			if !errors.As(err, &appErr) || appErr.Code != apperrors.ErrInternalServer.Code || appErr.StatusCode != 500 {
				t.Fatalf("expected internal error, got %v", err)
			}
		})
	}
}

func serializedInput(t *testing.T, input any, command CommandType) *fingerprintEnvelope {
	t.Helper()
	envelope, err := serializeCommand(input)
	if err != nil || envelope == nil {
		t.Fatalf("serialize input: %v", err)
	}
	if envelope.Command != command {
		t.Fatalf("command = %q, want %q", envelope.Command, command)
	}
	return envelope
}

func decodeFingerprint(t *testing.T, data []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode fingerprint: %v", err)
	}
}

func inputQuantity(t *testing.T, text string) quantity.Quantity {
	t.Helper()
	value, err := quantity.ParseQuantity(text)
	if err != nil {
		t.Fatal(err)
	}
	return *value
}

func inputDate(t *testing.T, text string) *domain.Date {
	t.Helper()
	value, err := domain.ParseDate(text)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func checkFingerprintDate(t *testing.T, got *string, want *domain.Date) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Fatalf("expected unknown date, got %q", *got)
		}
		return
	}
	if got == nil || *got != want.String() {
		t.Fatalf("date = %v, want %s", got, want.String())
	}
}
