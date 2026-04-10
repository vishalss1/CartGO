package util

import (
	"github.com/google/uuid"
)

// GenerateUUID centralizes UUID generation for the entire application.
// This allows uniform tracking, logging, or swapping of the UUID implementation if needed.
func GenerateUUID() uuid.UUID {
	return uuid.New()
}

// GenerateUUIDString returns a new UUID as a string.
func GenerateUUIDString() string {
	return uuid.New().String()
}
