package util

import (
	"context"
	"log"
	"net/http"
)

type contextKey string

const (
	CorrelationIDHeader contextKey = "X-Correlation-ID"
)

// GetCorrelationID extracts the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if val, ok := ctx.Value(CorrelationIDHeader).(string); ok {
		return val
	}
	return ""
}

// CorrelationIDMiddleware handles Correlation ID propagation
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(string(CorrelationIDHeader))
		if correlationID == "" {
			correlationID = GenerateUUIDString()
		}

		log.Printf("[CorrelationID] Request: %s | ID: %s", r.URL.Path, correlationID)

		// Set the header in the response so it's visible to the client
		w.Header().Set(string(CorrelationIDHeader), correlationID)

		// Set the ID in the request context
		ctx := context.WithValue(r.Context(), CorrelationIDHeader, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TraceRequest injects the correlation ID into the outgoing request headers
func TraceRequest(ctx context.Context, req *http.Request) {
	id := GetCorrelationID(ctx)
	if id != "" {
		req.Header.Set(string(CorrelationIDHeader), id)
	}
}
