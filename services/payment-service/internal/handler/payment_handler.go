package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vishalss1/CartGO/services/payment-service/internal/model"
	"github.com/vishalss1/CartGO/services/payment-service/internal/service"
)

type PaymentHandler struct {
	service service.PaymentService
	logger  *slog.Logger
	v       *validator.Validate
}

func NewPaymentHandler(s service.PaymentService, l *slog.Logger) *PaymentHandler {
	return &PaymentHandler{
		service: s,
		logger:  l,
		v:       validator.New(),
	}
}

func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req model.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		h.errorJSONResponse(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	payment, err := h.service.ProcessPayment(r.Context(), req.OrderID, req.Amount)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.jsonResponse(w, http.StatusOK, model.PaymentResponse{
		PaymentID: payment.ID,
		OrderID:   payment.OrderID,
		Status:    string(payment.Status),
		Amount:    payment.Amount,
	})
}

func (h *PaymentHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	errCode := "INTERNAL_ERROR"
	msg := "An unexpected error occurred"

	switch {
	case errors.Is(err, service.ErrAmountMismatch):
		code = http.StatusConflict
		errCode = "AMOUNT_MISMATCH"
		msg = err.Error()
	}

	h.logger.Error("handler error", "error", err, "method", r.Method, "route", r.URL.Path, "status", code)
	h.errorJSONResponse(w, code, errCode, msg)
}

func (h *PaymentHandler) jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (h *PaymentHandler) errorJSONResponse(w http.ResponseWriter, code int, errorCode string, message string) {
	h.jsonResponse(w, code, map[string]string{
		"error": message,
		"code":  errorCode,
	})
}
