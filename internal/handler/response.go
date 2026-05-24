package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"org-structure-api/internal/domain"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func respondWithError(w http.ResponseWriter, err error) {
	var statusCode int
	message := err.Error()

	switch {
	case errors.Is(err, domain.ErrNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidInput) || errors.Is(err, domain.ErrRequiredField):
		statusCode = http.StatusBadRequest
	case errors.Is(err, domain.ErrAlreadyExists) || errors.Is(err, domain.ErrCycleDetected):
		statusCode = http.StatusConflict
	default:
		statusCode = http.StatusInternalServerError
		message = "Internal Server Error"
	}
	respondWithJSON(w, statusCode, map[string]string{"error": message})
}
