package handler

import (
	"encoding/json"
	"net/http"

	"github.com/vishalss1/CartGO/services/user-service/internal/model"
	"github.com/vishalss1/CartGO/services/user-service/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.userService.Signup(r.Context(), req)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			ErrorJSONResponse(w, http.StatusConflict, err.Error())
			return
		}
		ErrorJSONResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusCreated, resp)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.userService.Login(r.Context(), req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			ErrorJSONResponse(w, http.StatusUnauthorized, err.Error())
			return
		}
		ErrorJSONResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.userService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			ErrorJSONResponse(w, http.StatusUnauthorized, err.Error())
			return
		}
		ErrorJSONResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.userService.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		ErrorJSONResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}
