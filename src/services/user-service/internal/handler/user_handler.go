package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/user-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/user-service/internal/model"
	"github.com/vishalss1/CartGO/services/user-service/internal/service"
)

type UserHandler struct {
	userService service.UserService
	validate    *validator.Validate
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.userService.Signup(r.Context(), req)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			util.WriteError(w, http.StatusConflict, err.Error())
			return
		}
		util.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.userService.Login(r.Context(), req)
	if err != nil {
		log.Printf("[UserHandler] Login failed for %s: %v", req.Email, err)
		if err == service.ErrInvalidCredentials {
			util.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}
		util.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.userService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			util.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}
		util.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.userService.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			util.WriteError(w, http.StatusConflict, "email already in use")
			return
		}
		util.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListAllUsers(r.Context())
	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSON(w, http.StatusOK, users)
}

func (h *UserHandler) AdminUpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req struct {
		Role string `json:"role" validate:"required,oneof=CUSTOMER WAREHOUSE_STAFF DELIVERY_PARTNER ADMIN SUPPORT_AGENT"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.userService.ChangeUserRole(r.Context(), userID, req.Role)
	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"message": "role updated successfully"})
}
