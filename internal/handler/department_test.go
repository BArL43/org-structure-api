package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"org-structure-api/internal/domain"
)

type mockDepartmentUseCase struct {
	onCreate func(context.Context, string, *int64) (*domain.Department, error)
	onGet    func(context.Context, int64, int, bool) (*domain.Department, error)
	onUpdate func(context.Context, int64, domain.DepartmentUpdate) (*domain.Department, error)
	onDelete func(context.Context, int64, string, *int64) error
}

func (m *mockDepartmentUseCase) Create(ctx context.Context, name string, parentID *int64) (*domain.Department, error) {
	if m.onCreate == nil {
		return nil, nil
	}
	return m.onCreate(ctx, name, parentID)
}
func (m *mockDepartmentUseCase) GetByID(ctx context.Context, id int64, depth int, includeEmployees bool) (*domain.Department, error) {
	if m.onGet == nil {
		return nil, nil
	}
	return m.onGet(ctx, id, depth, includeEmployees)
}
func (m *mockDepartmentUseCase) Update(ctx context.Context, id int64, update domain.DepartmentUpdate) (*domain.Department, error) {
	if m.onUpdate == nil {
		return nil, nil
	}
	return m.onUpdate(ctx, id, update)
}
func (m *mockDepartmentUseCase) Delete(ctx context.Context, id int64, mode string, reassignToID *int64) error {
	if m.onDelete == nil {
		return nil
	}
	return m.onDelete(ctx, id, mode, reassignToID)
}

func TestDepartmentHandlerCreateSuccess(t *testing.T) {
	mockUC := &mockDepartmentUseCase{onCreate: func(_ context.Context, name string, parentID *int64) (*domain.Department, error) {
		if parentID != nil {
			t.Fatalf("expected nil parent")
		}
		return &domain.Department{ID: 1, Name: name}, nil
	}}
	h := NewDepartmentHandler(mockUC)
	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewBufferString(`{"name":"Backend"}`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d", rr.Code)
	}
	var resp domain.Department
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.ID != 1 || resp.Name != "Backend" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestDepartmentHandlerRejectsUnknownJSONField(t *testing.T) {
	called := false
	mockUC := &mockDepartmentUseCase{onCreate: func(context.Context, string, *int64) (*domain.Department, error) {
		called = true
		return nil, nil
	}}
	h := NewDepartmentHandler(mockUC)
	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewBufferString(`{"name":"Backend","unexpected":true}`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
	if called {
		t.Fatal("use case must not be called")
	}
}

func TestDepartmentHandlerGetByIDParsesQuery(t *testing.T) {
	mockUC := &mockDepartmentUseCase{onGet: func(_ context.Context, id int64, depth int, include bool) (*domain.Department, error) {
		if id != 42 || depth != 3 || include {
			t.Fatalf("unexpected args: id=%d depth=%d include=%v", id, depth, include)
		}
		return &domain.Department{ID: id, Name: "Platform"}, nil
	}}
	h := NewDepartmentHandler(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/departments/42?depth=3&include_employees=false", nil)
	req.SetPathValue("id", "42")
	rr := httptest.NewRecorder()
	h.GetByID(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestDepartmentHandlerUpdateSupportsNullParent(t *testing.T) {
	mockUC := &mockDepartmentUseCase{onUpdate: func(_ context.Context, id int64, update domain.DepartmentUpdate) (*domain.Department, error) {
		if id != 7 || !update.ParentIDSet || update.ParentID != nil {
			t.Fatalf("unexpected update: %+v", update)
		}
		return &domain.Department{ID: id, Name: "Backend"}, nil
	}}
	h := NewDepartmentHandler(mockUC)
	req := httptest.NewRequest(http.MethodPatch, "/departments/7", bytes.NewBufferString(`{"parent_id":null}`))
	req.SetPathValue("id", "7")
	rr := httptest.NewRecorder()
	h.Update(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestDepartmentHandlerDeleteRequiresMode(t *testing.T) {
	h := NewDepartmentHandler(&mockDepartmentUseCase{})
	req := httptest.NewRequest(http.MethodDelete, "/departments/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	h.Delete(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}
