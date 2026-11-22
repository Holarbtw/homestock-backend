package http

import (
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"testing"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
)

func TestToCreateItemInput(t *testing.T) {
	input, err := toCreateItemInput(CreateItemRequest{Name: "Milk", Category: "FOOD", Unit: "ml"})
	if err != nil {
		t.Fatal(err)
	}
	if input.Name != "Milk" || input.Category != domain.Food || input.Unit != domain.Milliliter ||
		input.MinimumQuantity != nil || input.PackageQuantity != nil || input.ExternalImageURL != nil || input.FirstBatch != nil {
		t.Fatalf("unexpected input for a card without a batch: %+v", input)
	}

	var req CreateItemRequest
	if err := json.Unmarshal([]byte(`{
		"name":"Milk", "category":"FOOD", "unit":"ml", "minimum_quantity":"0",
		"external_image_url":"https://example.test/milk.png",
		"first_batch":{"quantity":"12.005", "purchased_on":"2026-11-21"}
	}`), &req); err != nil {
		t.Fatal(err)
	}
	input, err = toCreateItemInput(req)
	if err != nil {
		t.Fatal(err)
	}
	if input.MinimumQuantity == nil || input.MinimumQuantity.String() != "0.000" || input.PackageQuantity != nil {
		t.Fatalf("unexpected optional quantities: %+v", input)
	}
	if input.ExternalImageURL == nil || input.ExternalImageURL.String() != "https://example.test/milk.png" {
		t.Fatalf("unexpected external image URL: %v", input.ExternalImageURL)
	}
	batch := input.FirstBatch
	if batch == nil || batch.Quantity.String() != "12.005" || batch.PurchasedOn == nil ||
		batch.PurchasedOn.String() != "2026-11-21" || batch.ExpiresOn != nil {
		t.Fatalf("unexpected first batch: %+v", batch)
	}
}

func TestToCreateItemInputCollectsParseErrors(t *testing.T) {
	var req CreateItemRequest
	if err := json.Unmarshal([]byte(`{
		"name":"Milk", "category":"FOOD", "unit":"ml", "minimum_quantity":"none",
		"package_quantity":"1.0031", "external_image_url":"https://example.test/%zz",
		"first_batch":{"quantity":"oops", "purchased_on":"2026-02-30"}
	}`), &req); err != nil {
		t.Fatal(err)
	}
	input, err := toCreateItemInput(req)
	var customErr *apperrors.CustomError
	if input != nil || !errors.As(err, &customErr) || customErr.StatusCode != http.StatusBadRequest || customErr.Code != domain.ErrInvalidItemCreate.Code {
		t.Fatalf("expected invalid item create error, got input %+v and error %v", input, err)
	}
	got := make(map[string]string, len(customErr.Fields))
	for _, field := range customErr.Fields {
		got[field.Field] = field.Code
	}
	want := map[string]string{
		"minimum_quantity":         apperrors.ErrInvalidQuantityFormat.Code,
		"package_quantity":         apperrors.ErrQuantityPrecisionExceeded.Code,
		"external_image_url":       domain.ErrInvalidExternalImageURL.Code,
		"first_batch.quantity":     apperrors.ErrInvalidQuantityFormat.Code,
		"first_batch.purchased_on": domain.ErrInvalidDate.Code,
	}
	if len(customErr.Fields) != len(want) || !maps.Equal(got, want) {
		t.Fatalf("unexpected field errors: %+v", customErr.Fields)
	}
}
