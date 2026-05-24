package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"org-structure-api/internal/domain"
)

type mockDepartmentUseCase struct {
	onCreate func(name string, parentID *int64) (*domain.Department, error)
}

func (m *mockDepartmentUseCase) Create(name string, parentID *int64) (*domain.Department, error) {
	return m.onCreate(name, parentID)
}

func (m *mockDepartmentUseCase) GetByID(id int64, depth int, includeEmployees bool) (*domain.Department, error) {
	return nil, nil
}

func (m *mockDepartmentUseCase) Update(id int64, name *string, parentID *int64) (*domain.Department, error) {
	return nil, nil
}

func (m *mockDepartmentUseCase) Delete(id int64, mode string, reassignToID *int64) error {
	return nil
}

func TestDepartmentHandler_Create_Success(t *testing.T) {
	mockUC := &mockDepartmentUseCase{
		onCreate: func(name string, parentID *int64) (*domain.Department, error) {
			return &domain.Department{
				ID:   1,
				Name: name,
			}, nil
		},
	}

	h := NewDepartmentHandler(mockUC)

	body := []byte(`{"name": "Backend"}`)
	req, err := http.NewRequest(http.MethodPost, "/departments/", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var resp domain.Department
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Name != "Backend" || resp.ID != 1 {
		t.Errorf("handler returned unexpected body: got name=%s, id=%d", resp.Name, resp.ID)
	}
}

func TestDepartmentHandler_Create_ValidationError(t *testing.T) {
	mockUC := &mockDepartmentUseCase{
		onCreate: func(name string, parentID *int64) (*domain.Department, error) {
			return nil, domain.ErrInvalidInput
		},
	}

	h := NewDepartmentHandler(mockUC)

	body := []byte(`{"name": ""}`)
	req, err := http.NewRequest(http.MethodPost, "/departments/", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
