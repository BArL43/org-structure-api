package usecase

import (
	"context"
	"strings"
	"time"

	"org-structure-api/internal/domain"
)

type employeeUseCase struct {
	empRepo  domain.EmployeeRepository
	deptRepo domain.DepartmentRepository
}

func NewEmployeeUseCase(empRepo domain.EmployeeRepository, deptRepo domain.DepartmentRepository) domain.EmployeeUseCase {
	return &employeeUseCase{empRepo: empRepo, deptRepo: deptRepo}
}

func (u *employeeUseCase) Create(ctx context.Context, deptID int64, fullName, position string, hiredAt *time.Time) (*domain.Employee, error) {
	if deptID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	fullName = strings.TrimSpace(fullName)
	position = strings.TrimSpace(position)
	if !validTextField(fullName) || !validTextField(position) {
		return nil, domain.ErrInvalidInput
	}

	deptExists, err := u.deptRepo.Exists(ctx, deptID)
	if err != nil {
		return nil, err
	}
	if !deptExists {
		return nil, domain.ErrNotFound
	}

	emp := &domain.Employee{DepartmentID: deptID, FullName: fullName, Position: position, HiredAt: hiredAt}
	if err := u.empRepo.Create(ctx, emp); err != nil {
		return nil, err
	}
	return emp, nil
}
