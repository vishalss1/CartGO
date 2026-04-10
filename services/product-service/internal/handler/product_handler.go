package handler

import (
	"encoding/json"
	"errors"
	"log"
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
	v       *validator.Validate
}

func NewProductHandler(s service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: s,
		v:       validator.New(),
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		log.Printf("[ProductHandler] invalid content type: %s | method: %s | route: %s | status: %d", r.Header.Get("Content-Type"), r.Method, r.URL.Path, http.StatusUnsupportedMediaType)
		ErrorJSONResponse(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProductHandler] failed to decode request: %v | method: %s | route: %s | status: %d", err, r.Method, r.URL.Path, http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.v.Struct(req); err != nil {
		log.Printf("[ProductHandler] validation failed: %v | method: %s | route: %s | status: %d", err, r.Method, r.URL.Path, http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Validation failed")
		return
	}

	p, err := h.service.CreateProduct(r.Context(), &req)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	log.Printf("[ProductHandler] product created: %s | method: %s | route: %s | status: %d", p.ID, r.Method, r.URL.Path, http.StatusCreated)
	JSONResponse(w, http.StatusCreated, p)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[ProductHandler] invalid uuid: %s | method: %s | route: %s | status: %d", id, r.Method, r.URL.Path, http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p, err := h.service.GetProduct(r.Context(), parsedID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	log.Printf("[ProductHandler] product retrieved: %s | method: %s | route: %s | status: %d", id, r.Method, r.URL.Path, http.StatusOK)
	JSONResponse(w, http.StatusOK, p)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	if limit < 0 || offset < 0 {
		log.Printf("[ProductHandler] invalid pagination params | method: %s | route: %s | status: %d", r.Method, r.URL.Path, http.StatusBadRequest)
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

	log.Printf("[ProductHandler] products listed: count=%d | method: %s | route: %s | status: %d", len(products), r.Method, r.URL.Path, http.StatusOK)
	JSONResponse(w, http.StatusOK, products)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		log.Printf("[ProductHandler] invalid content type: %s | method: %s | route: %s | status: %d", r.Header.Get("Content-Type"), r.Method, r.URL.Path, http.StatusUnsupportedMediaType)
		ErrorJSONResponse(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	id := chi.URLParam(r, "id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[ProductHandler] invalid uuid: %s | method: %s | route: %s | status: %d", id, r.Method, r.URL.Path, http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProductHandler] failed to decode request: %v | method: %s | route: %s | status: %d", err, r.Method, r.URL.Path, http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Edge case check: At least one field must be provided
	if req.Name == "" && req.Description == "" && req.Price == nil && req.Category == "" {
		log.Printf("[ProductHandler] empty update request: %s | method: %s | route: %s | status: %d", id, r.Method, r.URL.Path, http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "At least one valid field must be provided")
		return
	}

	if err := h.v.Struct(req); err != nil {
		log.Printf("[ProductHandler] validation failed: %v | method: %s | route: %s | status: %d", err, r.Method, r.URL.Path, http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Validation failed")
		return
	}

	p, err := h.service.UpdateProduct(r.Context(), parsedID, &req)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	log.Printf("[ProductHandler] product updated: %s | method: %s | route: %s | status: %d", id, r.Method, r.URL.Path, http.StatusOK)
	JSONResponse(w, http.StatusOK, p)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[ProductHandler] invalid uuid: %s | method: %s | route: %s | status: %d", id, r.Method, r.URL.Path, http.StatusBadRequest)
		ErrorJSONResponse(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	err = h.service.DeleteProduct(r.Context(), parsedID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	log.Printf("[ProductHandler] product deleted: %s | method: %s | route: %s | status: %d", id, r.Method, r.URL.Path, http.StatusOK)
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

	log.Printf("[ProductHandler] error: %v | method: %s | route: %s | status: %d", err, r.Method, r.URL.Path, code)
	ErrorJSONResponse(w, code, msg)
}
