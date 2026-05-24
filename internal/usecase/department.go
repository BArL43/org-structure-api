package usecase

import (
	"org-structure-api/internal/domain"
	"strings"
)

type departmentUseCase struct {
	repo domain.DepartmentRepository
}

func NewDepartmentUseCase(repo domain.DepartmentRepository) domain.DepartmentUseCase {
	return &departmentUseCase{repo: repo}
}

func (u *departmentUseCase) Create(name string, parentID *int64) (*domain.Department, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 200 {
		return nil, domain.ErrInvalidInput
	}

	if parentID != nil {
		exists, err := u.repo.Exists(*parentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.ErrNotFound
		}
	}

	hasDuplicate, err := u.repo.HasChildWithSameName(parentID, name)
	if err != nil {
		return nil, err
	}
	if hasDuplicate {
		return nil, domain.ErrAlreadyExists
	}

	dept := &domain.Department{
		Name:     name,
		ParentID: parentID,
	}

	if err := u.repo.Create(dept); err != nil {
		return nil, err
	}

	return dept, nil
}

func (u *departmentUseCase) GetByID(id int64, depth int, includeEmployee bool) (*domain.Department, error) {
	if depth < 1 || depth > 5 {
		return nil, domain.ErrInvalidInput
	}

	dept, err := u.repo.GetByID(id, depth, includeEmployee)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, domain.ErrNotFound
	}

	return dept, nil
}

func (u *departmentUseCase) Update(id int64, name *string, parentID *int64) (*domain.Department, error) {
	exists, err := u.repo.Exists(id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrNotFound
	}

	updatedDept := &domain.Department{ID: id}

	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if trimmedName == "" || len([]rune(trimmedName)) > 200 {
			return nil, domain.ErrInvalidInput
		}

		updatedDept.Name = trimmedName
	}

	if parentID != nil {
		if *parentID == id {
			return nil, domain.ErrInvalidInput
		}

		parentExists, err := u.repo.Exists(*parentID)
		if err != nil {
			return nil, err
		}
		if !parentExists {
			return nil, domain.ErrNotFound
		}

		isLoop, err := u.repo.IsAncesstor(id, *parentID)
		if err != nil {
			return nil, err
		}
		if isLoop {
			return nil, domain.ErrCycleDetected
		}

		updatedDept.ParentID = parentID
	}

	if err := u.repo.Update(updatedDept); err != nil {
		return nil, err
	}

	return u.repo.GetByID(id, 1, false)
}

func (u *departmentUseCase) Delete(id int64, mode string, reassignID *int64) error {
	exists, err := u.repo.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}

	switch mode {
	case "cascade":
		return u.repo.DeleteCascade(id)

	case "reassign":
		if reassignID == nil {
			return domain.ErrRequiredField
		}

		targetExists, err := u.repo.Exists(*reassignID)
		if err != nil {
			return err
		}
		if !targetExists {
			return domain.ErrNotFound
		}

		if *reassignID == id {
			return domain.ErrInvalidInput
		}
		return u.repo.DeleteAndReassign(id, *reassignID)
	default:
		return domain.ErrInvalidInput
	}

}
