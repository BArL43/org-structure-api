package usecase

import (
	"context"
	"strings"

	"org-structure-api/internal/domain"
)

type departmentUseCase struct {
	repo domain.DepartmentRepository
}

func NewDepartmentUseCase(repo domain.DepartmentRepository) domain.DepartmentUseCase {
	return &departmentUseCase{repo: repo}
}

func (u *departmentUseCase) Create(ctx context.Context, name string, parentID *int64) (*domain.Department, error) {
	name = strings.TrimSpace(name)
	if !validTextField(name) {
		return nil, domain.ErrInvalidInput
	}
	if parentID != nil {
		if *parentID <= 0 {
			return nil, domain.ErrInvalidInput
		}
		exists, err := u.repo.Exists(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.ErrNotFound
		}
	}

	hasDuplicate, err := u.repo.HasChildWithSameName(ctx, parentID, name, nil)
	if err != nil {
		return nil, err
	}
	if hasDuplicate {
		return nil, domain.ErrAlreadyExists
	}

	dept := &domain.Department{Name: name, ParentID: parentID}
	if err := u.repo.Create(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (u *departmentUseCase) GetByID(ctx context.Context, id int64, depth int, includeEmployees bool) (*domain.Department, error) {
	if id <= 0 || depth < 1 || depth > 5 {
		return nil, domain.ErrInvalidInput
	}
	dept, err := u.repo.GetByID(ctx, id, depth, includeEmployees)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, domain.ErrNotFound
	}
	return dept, nil
}

func (u *departmentUseCase) Update(ctx context.Context, id int64, update domain.DepartmentUpdate) (*domain.Department, error) {
	if id <= 0 || (update.Name == nil && !update.ParentIDSet) {
		return nil, domain.ErrInvalidInput
	}

	current, err := u.repo.GetByID(ctx, id, 0, false)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, domain.ErrNotFound
	}

	effectiveName := current.Name
	if update.Name != nil {
		trimmedName := strings.TrimSpace(*update.Name)
		if !validTextField(trimmedName) {
			return nil, domain.ErrInvalidInput
		}
		effectiveName = trimmedName
		update.Name = &effectiveName
	}

	effectiveParentID := current.ParentID
	if update.ParentIDSet {
		if update.ParentID != nil {
			if *update.ParentID <= 0 || *update.ParentID == id {
				return nil, domain.ErrInvalidInput
			}
			parentExists, err := u.repo.Exists(ctx, *update.ParentID)
			if err != nil {
				return nil, err
			}
			if !parentExists {
				return nil, domain.ErrNotFound
			}
			isLoop, err := u.repo.IsAncestor(ctx, id, *update.ParentID)
			if err != nil {
				return nil, err
			}
			if isLoop {
				return nil, domain.ErrCycleDetected
			}
		}
		effectiveParentID = update.ParentID
	}

	excludeID := id
	hasDuplicate, err := u.repo.HasChildWithSameName(ctx, effectiveParentID, effectiveName, &excludeID)
	if err != nil {
		return nil, err
	}
	if hasDuplicate {
		return nil, domain.ErrAlreadyExists
	}

	if err := u.repo.Update(ctx, id, update); err != nil {
		return nil, err
	}
	updated, err := u.repo.GetByID(ctx, id, 1, false)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, domain.ErrNotFound
	}
	return updated, nil
}

func (u *departmentUseCase) Delete(ctx context.Context, id int64, mode string, reassignID *int64) error {
	if id <= 0 {
		return domain.ErrInvalidInput
	}
	exists, err := u.repo.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}

	switch mode {
	case "cascade":
		return u.repo.DeleteCascade(ctx, id)
	case "reassign":
		if reassignID == nil {
			return domain.ErrRequiredField
		}
		if *reassignID <= 0 || *reassignID == id {
			return domain.ErrInvalidInput
		}
		targetExists, err := u.repo.Exists(ctx, *reassignID)
		if err != nil {
			return err
		}
		if !targetExists {
			return domain.ErrNotFound
		}
		return u.repo.DeleteAndReassign(ctx, id, *reassignID)
	default:
		return domain.ErrInvalidInput
	}
}

func validTextField(value string) bool {
	length := len([]rune(value))
	return length > 0 && length <= 200
}
