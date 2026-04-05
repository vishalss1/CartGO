package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// In a real system, we'd log this error
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func ErrorJSONResponse(w http.ResponseWriter, statusCode int, errorCode string, message string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
	}
	JSONResponse(w, statusCode, resp)
}
