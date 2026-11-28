package http

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/Holarbtw/homestock-backend/pkg/httpresponse"
	pkgJWT "github.com/Holarbtw/homestock-backend/pkg/jwt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type StockHandler struct {
	stockService ports.StockService
}

func NewStockHandler(stockService ports.StockService) *StockHandler {
	return &StockHandler{
		stockService: stockService,
	}
}

// @Summary Replenish an inventory item
// @Description Adds a batch to an active item and returns the updated item, the new batch and its REPLENISH operation.
// @Description Send a positive quantity as a decimal string with up to three fractional digits. The optional purchased_on and expires_on fields use YYYY-MM-DD. The purchase date must be today or earlier.
// @Description Retry the same request with the same Idempotency-Key to receive the saved result with status 201. Reusing the key for a different request returns 409.
// @Tags Items
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param item_id path string true "Item UUID" format(uuid)
// @Param Idempotency-Key header string true "Client-generated request key, up to 200 bytes"
// @Param input body AddBatchRequest true "Batch data"
// @Success 201 {object} DataResponse[StockChangeResponse]
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_ADD_BATCH, INVALID_BATCH_QUANTITY, PURCHASE_DATE_IN_FUTURE, INVALID_IDEMPOTENCY_KEY"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 404 {object} apperrors.CustomError "ITEM_NOT_FOUND, SPACE_NOT_FOUND"
// @Failure 409 {object} apperrors.CustomError "IDEMPOTENCY_KEY_CONFLICT"
// @Failure 415 {object} apperrors.CustomError "UNSUPPORTED_MEDIA_TYPE"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /items/{item_id}/batches [post]
func (h *StockHandler) Replenish(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		apperrors.RespondWithError(w, r, apperrors.ErrUnsupportedMediaType)
		return
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "item_id"))
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest)
		return
	}
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	var addBatchRequest *AddBatchRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&addBatchRequest); err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	if addBatchRequest == nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest)
		return
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest)
		return
	}

	input, err := toAddBatchInput(*addBatchRequest)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}
	input.ItemID = itemID

	stock, batches, operation, err := h.stockService.Replenish(r.Context(), idempotencyKey, input, userID)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	if operation == nil {
		err := apperrors.ErrInternalServer.WithInternal(errors.New("replenishment result is missing the stock operation"))
		apperrors.RespondWithError(w, r, err)
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, DataResponse[StockChangeResponse]{
		Data: toStockChangeResponse(stock, batches, *operation),
	})
}
