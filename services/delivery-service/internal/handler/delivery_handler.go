package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/model"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/service"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/util"
)

type DeliveryHandler struct {
	service service.DeliveryService
	logger  *slog.Logger
	v       *validator.Validate
}

func NewDeliveryHandler(s service.DeliveryService, l *slog.Logger) *DeliveryHandler {
	return &DeliveryHandler{
		service: s,
		logger:  l,
		v:       validator.New(),
	}
}

func (h *DeliveryHandler) CreateDelivery(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		util.ErrorJSONResponse(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	delivery, err := h.service.CreateDelivery(r.Context(), req.OrderID, req.DeliveryAddress)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.JSONResponse(w, http.StatusCreated, delivery)
}

func (h *DeliveryHandler) UpdateDeliveryStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid delivery ID format")
		return
	}

	var req model.UpdateDeliveryStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		util.ErrorJSONResponse(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	err = h.service.UpdateDeliveryStatus(r.Context(), id, req.Status, req.DeliveryPersonID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.JSONResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DeliveryHandler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid delivery ID format")
		return
	}

	delivery, err := h.service.GetDelivery(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.JSONResponse(w, http.StatusOK, delivery)
}

func (h *DeliveryHandler) ListDeliveriesByPartner(w http.ResponseWriter, r *http.Request) {
	partnerIDStr := chi.URLParam(r, "partner_id")
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		util.ErrorJSONResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid partner ID format")
		return
	}

	deliveries, err := h.service.ListDeliveriesByPartner(r.Context(), partnerID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.JSONResponse(w, http.StatusOK, deliveries)
}

func (h *DeliveryHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	errCode := "INTERNAL_ERROR"
	msg := "An unexpected error occurred"

	// Simplified error mapping for delivery service
	h.logger.Error("handler error", "error", err, "method", r.Method, "route", r.URL.Path, "status", code)
	util.ErrorJSONResponse(w, code, errCode, msg)
}
