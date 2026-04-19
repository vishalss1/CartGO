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
	"github.com/vishalss1/CartGO/services/payment-service/internal/model"
	"github.com/vishalss1/CartGO/services/payment-service/internal/service"
)

type PaymentHandler struct {
	service service.PaymentService
	v       *validator.Validate
}

func NewPaymentHandler(s service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: s,
		v:       validator.New(),
	}
}

func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req model.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	payment, err := h.service.ProcessPayment(r.Context(), req.OrderID, req.Amount)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, model.PaymentResponse{
		PaymentID: payment.ID,
		OrderID:   payment.OrderID,
		Status:    string(payment.Status),
		Amount:    payment.Amount,
	})
}

func (h *PaymentHandler) GetPaymentByOrderID(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid order ID format")
		return
	}

	payment, err := h.service.GetPaymentByOrderID(r.Context(), orderID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	if payment == nil {
		util.WriteError(w, http.StatusNotFound, "Payment record not found for this order")
		return
	}

	util.WriteJSON(w, http.StatusOK, model.PaymentResponse{
		PaymentID: payment.ID,
		OrderID:   payment.OrderID,
		Status:    string(payment.Status),
		Amount:    payment.Amount,
	})
}

func (h *PaymentHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	msg := "An unexpected error occurred"

	switch {
	case errors.Is(err, service.ErrAmountMismatch):
		code = http.StatusConflict
		msg = err.Error()
	}

	log.Printf("[PaymentHandler] error: %v | method: %s | route: %s | status: %d", err, r.Method, r.URL.Path, code)
	util.WriteError(w, code, msg)
}

