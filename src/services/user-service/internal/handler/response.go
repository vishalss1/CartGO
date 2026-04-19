package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func JSONResponse(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func ErrorJSONResponse(w http.ResponseWriter, code int, message string) {
	JSONResponse(w, code, ErrorResponse{Error: message})
}
