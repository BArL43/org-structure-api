package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"org-structure-api/internal/domain"
)

type departmentHandler struct {
	useCase domain.DepartmentUseCase
}

func NewDepartmentHandler(useCase domain.DepartmentUseCase) *departmentHandler {
	return &departmentHandler{useCase: useCase}
}

func (h *departmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		ParentID *int64 `json:"parent_id"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		respondWithError(w, err)
		return
	}
	dept, err := h.useCase.Create(r.Context(), req.Name, req.ParentID)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, dept)
}

func (h *departmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondWithError(w, err)
		return
	}

	depth := 1
	if rawDepth := r.URL.Query().Get("depth"); rawDepth != "" {
		depth, err = strconv.Atoi(rawDepth)
		if err != nil {
			respondWithError(w, domain.ErrInvalidInput)
			return
		}
	}

	includeEmployees := true
	if rawInclude := r.URL.Query().Get("include_employees"); rawInclude != "" {
		includeEmployees, err = strconv.ParseBool(rawInclude)
		if err != nil {
			respondWithError(w, domain.ErrInvalidInput)
			return
		}
	}

	dept, err := h.useCase.GetByID(r.Context(), id, depth, includeEmployees)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, dept)
}

func (h *departmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondWithError(w, err)
		return
	}

	var req struct {
		Name     *string         `json:"name"`
		ParentID json.RawMessage `json:"parent_id"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		respondWithError(w, err)
		return
	}

	update := domain.DepartmentUpdate{Name: req.Name}
	if req.ParentID != nil {
		update.ParentIDSet = true
		if !bytes.Equal(bytes.TrimSpace(req.ParentID), []byte("null")) {
			var parentID int64
			if err := json.Unmarshal(req.ParentID, &parentID); err != nil || parentID <= 0 {
				respondWithError(w, domain.ErrInvalidInput)
				return
			}
			update.ParentID = &parentID
		}
	}

	dept, err := h.useCase.Update(r.Context(), id, update)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, dept)
}

func (h *departmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondWithError(w, err)
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		respondWithError(w, domain.ErrRequiredField)
		return
	}

	var reassignID *int64
	if rawID := r.URL.Query().Get("reassign_to_department_id"); rawID != "" {
		value, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || value <= 0 {
			respondWithError(w, domain.ErrInvalidInput)
			return
		}
		reassignID = &value
	}

	if err := h.useCase.Delete(r.Context(), id, mode, reassignID); err != nil {
		respondWithError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, domain.ErrInvalidInput
	}
	return id, nil
}
