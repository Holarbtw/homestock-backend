package http

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/Holarbtw/homestock-backend/pkg/httpresponse"
	pkgJWT "github.com/Holarbtw/homestock-backend/pkg/jwt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ItemHandler struct {
	itemService ports.ItemService
}

func NewItemHandler(itemService ports.ItemService) *ItemHandler {
	return &ItemHandler{
		itemService: itemService,
	}
}

// @Summary Get an inventory item
// @Description Returns an active item in the authenticated user's space with its total remaining quantity and batches.
// @Description The total includes expired batches. Items without batches return an empty batches array.
// @Tags Items
// @Security ApiKeyAuth
// @Produce json
// @Param item_id path string true "Item UUID" format(uuid)
// @Success 200 {object} DataResponse[ItemDetailsResponse]
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 404 {object} apperrors.CustomError "ITEM_NOT_FOUND, SPACE_NOT_FOUND"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /items/{item_id} [get]
func (h *ItemHandler) GetItem(w http.ResponseWriter, r *http.Request) {
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

	stock, batches, err := h.itemService.GetItem(r.Context(), userID, itemID)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	response := DataResponse[*ItemDetailsResponse]{Data: toItemDetailsResponse(stock, batches)}
	httpresponse.WriteJSON(w, http.StatusOK, response)
}

// @Summary Update an inventory item
// @Description Updates the name, brand, category or minimum quantity and returns the updated item.
// @Description Omitted fields keep their existing values. Explicit null clears brand or minimum_quantity. Name and category require a valid value when supplied. Send quantities as decimal strings.
// @Tags Items
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param item_id path string true "Item UUID" format(uuid)
// @Param input body UpdateItemRequest true "Fields to update"
// @Success 200 {object} DataResponse[ItemResponse]
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_ITEM_UPDATE, INVALID_QUANTITY_FORMAT, QUANTITY_PRECISION_EXCEEDED, QUANTITY_OUT_OF_RANGE"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 404 {object} apperrors.CustomError "ITEM_NOT_FOUND, SPACE_NOT_FOUND"
// @Failure 415 {object} apperrors.CustomError "UNSUPPORTED_MEDIA_TYPE"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /items/{item_id} [patch]
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
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

	var updateRequest *UpdateItemRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updateRequest); err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	if updateRequest == nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest)
		return
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest)
		return
	}

	input, err := toUpdateItemInput(*updateRequest)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}
	input.ItemID = itemID

	updated, err := h.itemService.UpdateItem(r.Context(), userID, input)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	response := DataResponse[ItemResponse]{Data: toItemResponse(updated)}
	httpresponse.WriteJSON(w, http.StatusOK, response)
}

// @Summary Delete an inventory item
// @Description Removes an item from active inventory and preserves its batches and operation history.
// @Description Items with remaining stock can be deleted. A repeat deletion returns 404.
// @Tags Items
// @Security ApiKeyAuth
// @Produce json
// @Param item_id path string true "Item UUID" format(uuid)
// @Success 204 "Item deleted successfully"
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 404 {object} apperrors.CustomError "ITEM_NOT_FOUND, SPACE_NOT_FOUND"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /items/{item_id} [delete]
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
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

	err = h.itemService.DeleteItem(r.Context(), userID, itemID)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary List inventory items
// @Description Returns a page of active items ordered by creation time and ID in descending order. Each item's total includes expired batches.
// @Description Use meta.next_offset for the next page. When the last page is reached, meta.has_more is false and meta.next_offset is null. Empty pages return an empty data array.
// @Tags Items
// @Security ApiKeyAuth
// @Produce json
// @Param limit query int false "Page size. Values above 100 are capped at 100" default(20) minimum(1)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Success 200 {object} PaginatedResponse[[]ItemResponse]
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /items [get]
func (h *ItemHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}

	limit, offset, err := parsePagination(r.URL.RawQuery)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	items, err := h.itemService.ListItems(r.Context(), userID, limit+1, offset)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	meta := PaginationMeta{Limit: limit, Offset: offset, HasMore: len(items) > limit}
	if meta.HasMore {
		items = items[:limit]
		nextOffset := offset + limit
		meta.NextOffset = &nextOffset
	}

	data := make([]ItemResponse, len(items))
	for i := range items {
		data[i] = toItemResponse(&items[i])
	}
	response := PaginatedResponse[[]ItemResponse]{Data: data, Meta: meta}
	httpresponse.WriteJSON(w, http.StatusOK, response)
}

// @Summary Create an inventory item
// @Description Creates an item with an optional first batch. Without a first batch, batches is empty and operation is omitted.
// @Description Retry the same request with the same Idempotency-Key to receive the saved result with status 201. Reusing the key for different input returns 409.
// @Tags Items
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "Client-generated request key, up to 200 bytes"
// @Param input body CreateItemRequest true "Item and optional first batch"
// @Success 201 {object} DataResponse[CreateItemResponse]
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_ITEM_CREATE"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 409 {object} apperrors.CustomError "IDEMPOTENCY_KEY_CONFLICT"
// @Failure 415 {object} apperrors.CustomError "UNSUPPORTED_MEDIA_TYPE"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /items [post]
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		apperrors.RespondWithError(w, r, apperrors.ErrUnsupportedMediaType)
		return
	}
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	var parsed *CreateItemRequest
	err = json.NewDecoder(r.Body).Decode(&parsed)
	if err != nil || parsed == nil{
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	input, err := toCreateItemInput(*parsed)
	if err != nil{
		apperrors.RespondWithError(w, r, err)
		return
	}
	stock, batches, operation, err := h.itemService.CreateItem(r.Context(), userID, idempotencyKey, input)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return 
	}

	response := toCreateItemResponse(stock, batches, operation)
	httpresponse.WriteJSON(w, http.StatusCreated, DataResponse[*CreateItemResponse]{
		Data: response,
	})
}

func parsePagination(rawQuery string) (limit, offset int, err error) {
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return 0, 0, apperrors.ErrInvalidRequest.WithInternal(err)
	}

	limit, offset = 20, 0

	if query.Has("limit") {
		limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil || limit <= 0 {
			return 0, 0, apperrors.ErrInvalidRequest
		}

		limit = min(limit, 100)
	}

	if query.Has("offset") {
		offset, err = strconv.Atoi(query.Get("offset"))
		if err != nil || offset < 0 {
			return 0, 0, apperrors.ErrInvalidRequest
		}
	}

	return limit, offset, nil
}
