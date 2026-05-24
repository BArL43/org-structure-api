package handler

import(
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

// POST /departments
func (h *departmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		ParentID *int64 `json:"parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, domain.ErrInvalidInput)
		return
	}

	dept, err := h.useCase.Create(req.Name, req.ParentID)
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, dept)
}

// GET /departments/{id}
func (h *departmentHandler) GetById(w http.ResponseWriter, r *http.Request) {

}

