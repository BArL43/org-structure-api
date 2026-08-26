package handler

import (
	"net/http"
	"time"

	"org-structure-api/internal/domain"
)

type EmployeeHandler struct {
	useCase domain.EmployeeUseCase
}

func NewEmployeeHandler(useCase domain.EmployeeUseCase) *EmployeeHandler {
	return &EmployeeHandler{useCase: useCase}
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	deptID, err := pathID(r)
	if err != nil {
		respondWithError(w, err)
		return
	}

	var req struct {
		FullName string     `json:"full_name"`
		Position string     `json:"position"`
		HiredAt  *time.Time `json:"hired_at"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		respondWithError(w, err)
		return
	}

	emp, err := h.useCase.Create(r.Context(), deptID, req.FullName, req.Position, req.HiredAt)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, emp)
}
