package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/model"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/service"
)

type DeliveryHandler struct {
	service service.DeliveryService
	v       *validator.Validate
}

func NewDeliveryHandler(s service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{
		service: s,
		v:       validator.New(),
	}
}

func (h *DeliveryHandler) CreateDelivery(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	delivery, err := h.service.CreateDelivery(r.Context(), req.OrderID, req.DeliveryAddress)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusCreated, delivery)
}

func (h *DeliveryHandler) UpdateDeliveryStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid delivery ID format")
		return
	}

	var req model.UpdateDeliveryStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.UpdateDeliveryStatus(r.Context(), id, req.Status, req.DeliveryPersonID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DeliveryHandler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid delivery ID format")
		return
	}

	delivery, err := h.service.GetDelivery(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, delivery)
}

func (h *DeliveryHandler) ListDeliveriesByPartner(w http.ResponseWriter, r *http.Request) {
	partnerIDStr := chi.URLParam(r, "partner_id")
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid partner ID format")
		return
	}

	deliveries, err := h.service.ListDeliveriesByPartner(r.Context(), partnerID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	util.WriteJSON(w, http.StatusOK, deliveries)
}

func (h *DeliveryHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	msg := "An unexpected error occurred"

	log.Printf("[DeliveryHandler] error: %v | method: %s | route: %s | status: %d", err, r.Method, r.URL.Path, code)
	util.WriteError(w, code, msg)
}
