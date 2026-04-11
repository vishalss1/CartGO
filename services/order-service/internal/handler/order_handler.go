package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/order-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/order-service/internal/model"
	"github.com/vishalss1/CartGO/services/order-service/internal/service"
	"github.com/vishalss1/CartGO/pkg/util"
)

type OrderHandler struct {
	service service.OrderService
	v       *validator.Validate
}

func NewOrderHandler(s service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: s,
		v:       validator.New(),
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Extract UserID from context (trusted source)
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		util.WriteError(w, http.StatusUnauthorized, "User identity not found or invalid")
		return
	}

	order, err := h.service.CreateOrder(r.Context(), userID, &req)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid order ID format")
		return
	}

	order, err := h.service.GetOrder(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) GetOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	orders, err := h.service.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	msg := "An unexpected error occurred"

	switch {
	case errors.Is(err, service.ErrNotFound):
		code = http.StatusNotFound
		msg = err.Error()
	case errors.Is(err, service.ErrInvalidOrder):
		code = http.StatusBadRequest
		msg = err.Error()
	case errors.Is(err, service.ErrStockFailed):
		code = http.StatusConflict
		msg = err.Error()
	case errors.Is(err, service.ErrPaymentFailed):
		code = http.StatusPaymentRequired
		msg = err.Error()
	}

	log.Printf("[OrderHandler] error: %v | method: %s | route: %s | status: %d", err, r.Method, r.URL.Path, code)
	util.WriteError(w, code, msg)
}
