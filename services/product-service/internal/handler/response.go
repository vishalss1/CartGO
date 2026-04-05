package handler

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *string     `json:"error"`
}

func JSONResponse(w http.ResponseWriter, code int, data interface{}) {
	success := code < 400
	resp := APIResponse{
		Success: success,
		Data:    data,
		Error:   nil,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func ErrorJSONResponse(w http.ResponseWriter, code int, message string) {
	resp := APIResponse{
		Success: false,
		Data:    nil,
		Error:   &message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}
