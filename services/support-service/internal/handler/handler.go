package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vishalss1/CartGO/services/support-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/support-service/internal/model"
	"github.com/vishalss1/CartGO/services/support-service/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) respondWithError(w http.ResponseWriter, code int, errCode string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(model.ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing User ID header")
		return
	}

	idempotencyKey := r.Header.Get("X-Idempotency-Key")
	if idempotencyKey == "" {
		h.respondWithError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing X-Idempotency-Key header")
		return
	}

	var req struct {
		Subject string `json:"subject"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	ticket, err := h.svc.CreateTicket(r.Context(), idempotencyKey, userID, req.Subject)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusCreated, ticket)
}

func (h *Handler) ListTickets(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetRole(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filters := make(map[string]interface{})
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}

	// Customers only view their OWN tickets
	if role != "ADMIN" && role != "SUPPORT_AGENT" {
		filters["customer_id"] = userID
	} else if agentID := r.URL.Query().Get("agent_id"); agentID != "" {
		filters["assigned_agent_id"] = agentID
	}

	tickets, err := h.svc.ListTickets(r.Context(), filters, limit, offset)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, tickets)
}

func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "id")
	ticket, err := h.svc.GetTicket(r.Context(), ticketID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, ticket)
}

func (h *Handler) AddMessage(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetRole(r.Context())

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	err := h.svc.AddMessage(r.Context(), ticketID, userID, role, req.Content)
	if err != nil {
		h.respondWithError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusCreated, map[string]string{"result": "success"})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetRole(r.Context())

	var req struct {
		Status model.TicketStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	err := h.svc.UpdateStatus(r.Context(), ticketID, req.Status, userID, role)
	if err != nil {
		h.respondWithError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (h *Handler) AssignTicket(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetRole(r.Context())

	var req struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	err := h.svc.AssignTicket(r.Context(), ticketID, req.AgentID, userID, role)
	if err != nil {
		h.respondWithError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetRole(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	messages, err := h.svc.ListMessages(r.Context(), ticketID, userID, role, limit, offset)
	if err != nil {
		h.respondWithError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, messages)
}
