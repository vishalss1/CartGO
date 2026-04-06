package util

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func JSONResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func ErrorJSONResponse(w http.ResponseWriter, code int, errorCode string, message string) {
	JSONResponse(w, code, ErrorResponse{
		ErrorCode: errorCode,
		Message:   message,
	})
}
