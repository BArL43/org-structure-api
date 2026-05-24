package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"org-structure-api/internal/domain"
)

type EmployeeHandler struct {
	useCase domain.EmployeeUseCase
}

func NewEmployeeHandler(useCase domain.EmployeeUseCase) *EmployeeHandler {
	return &EmployeeHandler{useCase: useCase}
}

// POST department/{id}/employees/
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	deptIDStr := r.PathValue("id")
	deptID, err := strconv.ParseInt(deptIDStr, 10, 64)
	if err != nil {
		respondWithError(w, domain.ErrInvalidInput)
		return
	}

	var req struct {
		FullName string     `json:"full_name"`
		Position string     `json:"position"`
		HiredAt  *time.Time `json:"hired_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, domain.ErrInvalidInput)
		return
	}

	emp, err := h.useCase.Create(deptID, req.FullName, req.Position, req.HiredAt)
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, emp)
}
