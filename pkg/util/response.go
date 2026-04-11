package util

import (
	"encoding/json"
	"net/http"
)

// APIResponse is the standard response structure for all microservices.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}

// WriteJSON sends a standardized JSON response.
func WriteJSON(w http.ResponseWriter, code int, data interface{}) {
	resp := APIResponse{
		Success: code < 400,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

// WriteError sends a standardized error response.
func WriteError(w http.ResponseWriter, code int, message string) {
	resp := APIResponse{
		Success: false,
		Error:   &message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}
