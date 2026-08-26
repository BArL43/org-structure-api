package usecase

import (
	"context"
	"errors"
	"testing"

	"org-structure-api/internal/domain"
)

type fakeDepartmentRepo struct {
	createFn        func(context.Context, *domain.Department) error
	getFn           func(context.Context, int64, int, bool) (*domain.Department, error)
	updateFn        func(context.Context, int64, domain.DepartmentUpdate) error
	deleteReFn      func(context.Context, int64, int64) error
	deleteCascadeFn func(context.Context, int64) error
	existsFn        func(context.Context, int64) (bool, error)
	duplicateFn     func(context.Context, *int64, string, *int64) (bool, error)
	ancestorFn      func(context.Context, int64, int64) (bool, error)
}

func (f *fakeDepartmentRepo) Create(ctx context.Context, d *domain.Department) error {
	if f.createFn != nil {
		return f.createFn(ctx, d)
	}
	return nil
}
func (f *fakeDepartmentRepo) GetByID(ctx context.Context, id int64, depth int, include bool) (*domain.Department, error) {
	if f.getFn != nil {
		return f.getFn(ctx, id, depth, include)
	}
	return nil, nil
}
func (f *fakeDepartmentRepo) Update(ctx context.Context, id int64, u domain.DepartmentUpdate) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, id, u)
	}
	return nil
}
func (f *fakeDepartmentRepo) DeleteAndReassign(ctx context.Context, id, target int64) error {
	if f.deleteReFn != nil {
		return f.deleteReFn(ctx, id, target)
	}
	return nil
}
func (f *fakeDepartmentRepo) DeleteCascade(ctx context.Context, id int64) error {
	if f.deleteCascadeFn != nil {
		return f.deleteCascadeFn(ctx, id)
	}
	return nil
}
func (f *fakeDepartmentRepo) Exists(ctx context.Context, id int64) (bool, error) {
	if f.existsFn != nil {
		return f.existsFn(ctx, id)
	}
	return true, nil
}
func (f *fakeDepartmentRepo) HasChildWithSameName(ctx context.Context, p *int64, n string, e *int64) (bool, error) {
	if f.duplicateFn != nil {
		return f.duplicateFn(ctx, p, n, e)
	}
	return false, nil
}
func (f *fakeDepartmentRepo) IsAncestor(ctx context.Context, a, d int64) (bool, error) {
	if f.ancestorFn != nil {
		return f.ancestorFn(ctx, a, d)
	}
	return false, nil
}

func TestDepartmentUseCaseUpdateNameOnlyKeepsParent(t *testing.T) {
	parentID := int64(5)
	updated := false
	repo := &fakeDepartmentRepo{
		getFn: func(_ context.Context, id int64, depth int, _ bool) (*domain.Department, error) {
			if depth == 0 {
				return &domain.Department{ID: id, Name: "Old", ParentID: &parentID}, nil
			}
			return &domain.Department{ID: id, Name: "New", ParentID: &parentID}, nil
		},
		duplicateFn: func(_ context.Context, p *int64, name string, exclude *int64) (bool, error) {
			if p == nil || *p != parentID || name != "New" || exclude == nil || *exclude != 10 {
				t.Fatalf("unexpected duplicate check")
			}
			return false, nil
		},
		updateFn: func(_ context.Context, id int64, u domain.DepartmentUpdate) error {
			updated = true
			if id != 10 || u.Name == nil || *u.Name != "New" || u.ParentIDSet {
				t.Fatalf("unexpected update: %+v", u)
			}
			return nil
		},
	}
	uc := NewDepartmentUseCase(repo)
	name := " New "
	_, err := uc.Update(context.Background(), 10, domain.DepartmentUpdate{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("update was not called")
	}
}

func TestDepartmentUseCaseUpdateDetectsCycle(t *testing.T) {
	newParent := int64(12)
	repo := &fakeDepartmentRepo{
		getFn:      func(context.Context, int64, int, bool) (*domain.Department, error) { return &domain.Department{ID: 10, Name: "Backend"}, nil },
		existsFn:   func(context.Context, int64) (bool, error) { return true, nil },
		ancestorFn: func(context.Context, int64, int64) (bool, error) { return true, nil },
	}
	uc := NewDepartmentUseCase(repo)
	_, err := uc.Update(context.Background(), 10, domain.DepartmentUpdate{ParentID: &newParent, ParentIDSet: true})
	if !errors.Is(err, domain.ErrCycleDetected) {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

func TestDepartmentUseCaseUpdateCanMoveDepartmentToRoot(t *testing.T) {
	oldParent := int64(2)
	repo := &fakeDepartmentRepo{
		getFn: func(_ context.Context, id int64, depth int, _ bool) (*domain.Department, error) {
			if depth == 0 {
				return &domain.Department{ID: id, Name: "Backend", ParentID: &oldParent}, nil
			}
			return &domain.Department{ID: id, Name: "Backend", ParentID: nil}, nil
		},
		duplicateFn: func(_ context.Context, p *int64, name string, _ *int64) (bool, error) {
			if p != nil {
				t.Fatal("expected root parent")
			}
			return false, nil
		},
		updateFn: func(_ context.Context, _ int64, u domain.DepartmentUpdate) error {
			if !u.ParentIDSet || u.ParentID != nil {
				t.Fatal("expected explicit null parent")
			}
			return nil
		},
	}
	uc := NewDepartmentUseCase(repo)
	_, err := uc.Update(context.Background(), 10, domain.DepartmentUpdate{ParentIDSet: true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDepartmentUseCaseCreateRejectsDuplicate(t *testing.T) {
	repo := &fakeDepartmentRepo{duplicateFn: func(context.Context, *int64, string, *int64) (bool, error) { return true, nil }}
	uc := NewDepartmentUseCase(repo)
	_, err := uc.Create(context.Background(), "Backend", nil)
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestDepartmentUseCaseDeleteReassignRequiresTarget(t *testing.T) {
	repo := &fakeDepartmentRepo{existsFn: func(context.Context, int64) (bool, error) { return true, nil }}
	uc := NewDepartmentUseCase(repo)
	err := uc.Delete(context.Background(), 1, "reassign", nil)
	if !errors.Is(err, domain.ErrRequiredField) {
		t.Fatalf("expected required field, got %v", err)
	}
}
