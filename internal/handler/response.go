package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"org-structure-api/internal/domain"
)

const maxRequestBodyBytes = 1 << 20

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func respondWithError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	message := "Internal Server Error"

	switch {
	case errors.Is(err, domain.ErrNotFound):
		statusCode = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrRequiredField):
		statusCode = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, domain.ErrAlreadyExists), errors.Is(err, domain.ErrCycleDetected):
		statusCode = http.StatusConflict
		message = err.Error()
	}
	respondWithJSON(w, statusCode, map[string]string{"error": message})
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return domain.ErrInvalidInput
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return domain.ErrInvalidInput
	}
	return nil
}
