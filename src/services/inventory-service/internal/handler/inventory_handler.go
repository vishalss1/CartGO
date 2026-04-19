package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/model"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/service"
)

type InventoryHandler struct {
	service service.InventoryService
	v       *validator.Validate
}

func NewInventoryHandler(s service.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		service: s,
		v:       validator.New(),
	}
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("[InventoryHandler] invalid uuid: %s | method: %s | route: %s | status: %d", idStr, r.Method, r.URL.Path, http.StatusBadRequest)
		util.WriteError(w, http.StatusBadRequest, "Invalid product ID format")
		return
	}

	inv, err := h.service.GetInventory(r.Context(), productID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, inv)
}

func (h *InventoryHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid product ID format")
		return
	}

	var req model.UpdateStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[InventoryHandler] decode error: %v", err)
		util.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	log.Printf("[InventoryHandler] adjusting stock for product %s by %d", productID, req.Adjustment)

	if err := h.v.Struct(req); err != nil {
		log.Printf("[InventoryHandler] validation error: %v", err)
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.AdjustStock(r.Context(), productID, req.Adjustment)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"status": "Stock adjusted successfully"})
}

func (h *InventoryHandler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid product ID format")
		return
	}

	var req model.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.ReserveStock(r.Context(), productID, req.OrderID, req.Quantity)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"status": "Stock reserved successfully"})
}

func (h *InventoryHandler) ReleaseStock(w http.ResponseWriter, r *http.Request) {
	var req model.IdempotentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.ReleaseStock(r.Context(), req.OrderID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"status": "Stock released successfully"})
}

func (h *InventoryHandler) CommitStock(w http.ResponseWriter, r *http.Request) {
	var req model.IdempotentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.CommitStock(r.Context(), req.OrderID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"status": "Stock committed successfully"})
}

func (h *InventoryHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	msg := "An unexpected error occurred"

	switch {
	case errors.Is(err, service.ErrNotFound):
		code = http.StatusNotFound
		msg = err.Error()
	case errors.Is(err, service.ErrInsufficientStock):
		code = http.StatusConflict
		msg = err.Error()
	case errors.Is(err, service.ErrConflict):
		code = http.StatusConflict
		msg = err.Error()
	case errors.Is(err, service.ErrInvalidState):
		code = http.StatusBadRequest
		msg = err.Error()
	}

	log.Printf("[InventoryHandler] error: %v | method: %s | route: %s | status: %d", err, r.Method, r.URL.Path, code)
	util.WriteError(w, code, msg)
}
