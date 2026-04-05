package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/services/product-service/internal/model"
	"github.com/vishalss1/CartGO/services/product-service/internal/service"
)

type ProductHandler struct {
	service service.ProductService
	logger  *slog.Logger
	v       *validator.Validate
}

func NewProductHandler(s service.ProductService, l *slog.Logger) *ProductHandler {
	return &ProductHandler{
		service: s,
		logger:  l,
		v:       validator.New(),
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		h.logger.Error("invalid content type", "method", r.Method, "route", r.URL.Path, "status", http.StatusUnsupportedMediaType)
		ErrorJSONResponse(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err, "method", r.Method, "route", r.URL.Path, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		h.logger.Error("validation failed", "error", err, "method", r.Method, "route", r.URL.Path, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Validation failed")
		return
	}

	p, err := h.service.CreateProduct(r.Context(), &req)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.logger.Info("product created", "id", p.ID, "method", r.Method, "route", r.URL.Path, "status", http.StatusCreated)
	JSONResponse(w, http.StatusCreated, p)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		h.logger.Error("invalid uuid", "id", id, "method", r.Method, "route", r.URL.Path, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.logger.Info("product retrieved", "id", id, "method", r.Method, "route", r.URL.Path, "status", http.StatusOK)
	JSONResponse(w, http.StatusOK, p)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	if limit < 0 || offset < 0 {
		h.logger.Error("invalid pagination params", "method", r.Method, "route", r.URL.Path, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Limit and offset must be non-negative")
		return
	}

	if limit == 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	filter := &model.ProductFilter{
		Category: query.Get("category"),
		Limit:    limit,
		Offset:   offset,
	}

	if minPriceStr := query.Get("min_price"); minPriceStr != "" {
		if val, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filter.MinPrice = &val
		}
	}
	if maxPriceStr := query.Get("max_price"); maxPriceStr != "" {
		if val, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filter.MaxPrice = &val
		}
	}

	products, err := h.service.ListProducts(r.Context(), filter)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Ensure empty array instead of null
	if products == nil {
		products = []*model.Product{}
	}

	h.logger.Info("products listed", "count", len(products), "method", r.Method, "route", r.URL.Path, "status", http.StatusOK)
	JSONResponse(w, http.StatusOK, products)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		h.logger.Error("invalid content type", "method", r.Method, "route", r.URL.Path, "status", http.StatusUnsupportedMediaType)
		ErrorJSONResponse(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		h.logger.Error("invalid uuid", "id", id, "method", r.Method, "route", r.URL.Path, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err, "method", r.Method, "route", r.URL.Path, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Edge case check: At least one field must be provided
	if req.Name == "" && req.Description == "" && req.Price == nil && req.Category == "" {
		h.logger.Error("empty update request", "id", id, "method", r.Method, "route", r.URL.Path, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "At least one valid field must be provided")
		return
	}

	if err := h.v.Struct(req); err != nil {
		h.logger.Error("validation failed", "error", err, "method", r.Method, "route", r.URL.Path, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Validation failed")
		return
	}

	p, err := h.service.UpdateProduct(r.Context(), id, &req)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.logger.Info("product updated", "id", id, "method", r.Method, "route", r.URL.Path, "status", http.StatusOK)
	JSONResponse(w, http.StatusOK, p)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		h.logger.Error("invalid uuid", "id", id, "method", r.Method, "route", r.URL.Path, "status", http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	err := h.service.DeleteProduct(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.logger.Info("product deleted", "id", id, "method", r.Method, "route", r.URL.Path, "status", http.StatusOK)
	JSONResponse(w, http.StatusOK, nil)
}

func (h *ProductHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	msg := "Internal server error"

	if errors.Is(err, service.ErrNotFound) {
		code = http.StatusNotFound
		msg = err.Error()
	} else if errors.Is(err, service.ErrInvalidInput) {
		code = http.StatusBadRequest
		msg = err.Error()
	}

	h.logger.Error("handler error", "error", err, "method", r.Method, "route", r.URL.Path, "status", code)
	ErrorJSONResponse(w, code, msg)
}
