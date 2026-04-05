package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/model"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/service"
)

type InventoryHandler struct {
	service service.InventoryService
	logger  *slog.Logger
	v       *validator.Validate
}

func NewInventoryHandler(s service.InventoryService, l *slog.Logger) *InventoryHandler {
	return &InventoryHandler{
		service: s,
		logger:  l,
		v:       validator.New(),
	}
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "product_id")
	if _, err := uuid.Parse(productID); err != nil {
		h.logger.Error("invalid uuid", "id", productID, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid product ID format")
		return
	}

	inv, err := h.service.GetInventory(r.Context(), productID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSONResponse(w, http.StatusOK, inv)
}

func (h *InventoryHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "product_id")
	if _, err := uuid.Parse(productID); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid product ID format")
		return
	}

	var req model.UpdateStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	err := h.service.AdjustStock(r.Context(), productID, req.TotalStock)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "Stock updated successfully"})
}

func (h *InventoryHandler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "product_id")
	if _, err := uuid.Parse(productID); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid product ID format")
		return
	}

	var req model.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	err := h.service.ReserveStock(r.Context(), productID, req.OrderID.String(), req.Quantity)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "Stock reserved successfully"})
}

func (h *InventoryHandler) ReleaseStock(w http.ResponseWriter, r *http.Request) {
	var req model.IdempotentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	err := h.service.ReleaseStock(r.Context(), req.OrderID.String())
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "Stock released successfully"})
}

func (h *InventoryHandler) CommitStock(w http.ResponseWriter, r *http.Request) {
	var req model.IdempotentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	err := h.service.CommitStock(r.Context(), req.OrderID.String())
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "Stock committed successfully"})
}

func (h *InventoryHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	errCode := "INTERNAL_ERROR"
	msg := "An unexpected error occurred"

	switch {
	case errors.Is(err, service.ErrNotFound):
		code = http.StatusNotFound
		errCode = "NOT_FOUND"
		msg = err.Error()
	case errors.Is(err, service.ErrInsufficientStock):
		code = http.StatusConflict
		errCode = "INSUFFICIENT_STOCK"
		msg = err.Error()
	case errors.Is(err, service.ErrConflict):
		code = http.StatusConflict
		errCode = "CONFLICT"
		msg = err.Error()
	case errors.Is(err, service.ErrInvalidState):
		code = http.StatusBadRequest
		errCode = "INVALID_STATE"
		msg = err.Error()
	}

	h.logger.Error("handler error", "error", err, "method", r.Method, "route", r.URL.Path, "status", code)
	ErrorJSONResponse(w, code, errCode, msg)
}
