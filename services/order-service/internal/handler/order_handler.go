package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/order-service/internal/model"
	"github.com/vishalss1/CartGO/services/order-service/internal/service"
)

type OrderHandler struct {
	service service.OrderService
	logger  *slog.Logger
	v       *validator.Validate
}

func NewOrderHandler(s service.OrderService, l *slog.Logger) *OrderHandler {
	return &OrderHandler{
		service: s,
		logger:  l,
		v:       validator.New(),
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	order, err := h.service.CreateOrder(r.Context(), &req)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSONResponse(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid order ID format")
		return
	}

	order, err := h.service.GetOrder(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSONResponse(w, http.StatusOK, order)
}

func (h *OrderHandler) GetOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID format")
		return
	}

	orders, err := h.service.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSONResponse(w, http.StatusOK, orders)
}

func (h *OrderHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	errCode := "INTERNAL_ERROR"
	msg := "An unexpected error occurred"

	switch {
	case errors.Is(err, service.ErrNotFound):
		code = http.StatusNotFound
		errCode = "NOT_FOUND"
		msg = err.Error()
	case errors.Is(err, service.ErrInvalidOrder):
		code = http.StatusBadRequest
		errCode = "INVALID_ORDER"
		msg = err.Error()
	case errors.Is(err, service.ErrStockFailed):
		code = http.StatusConflict
		errCode = "STOCK_FAILED"
		msg = err.Error()
	case errors.Is(err, service.ErrPaymentFailed):
		code = http.StatusPaymentRequired
		errCode = "PAYMENT_FAILED"
		msg = err.Error()
	}

	h.logger.Error("handler error", "error", err, "method", r.Method, "route", r.URL.Path, "status", code)
	ErrorJSONResponse(w, code, errCode, msg)
}
