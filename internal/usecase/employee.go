package usecase

import (
	"org-structure-api/internal/domain"
	"strings"
	"time"
)

type employeeUseCase struct {
	empRepo  domain.EmployeeRepository
	deptRepo domain.DepartmentRepository
}

func NewEmployeeUseCase(empRepo domain.EmployeeRepository, deptRepo domain.DepartmentRepository) domain.EmployeeUseCase {
	return &employeeUseCase{empRepo: empRepo, deptRepo: deptRepo}
}

func (u *employeeUseCase) Create(deptID int64, fullName, position string, hiredAt *time.Time) (*domain.Employee, error) {
	fullName = strings.TrimSpace(fullName)
	position = strings.TrimSpace(position)

	if fullName == "" || len([]rune(fullName)) > 200 {
		return nil, domain.ErrInvalidInput
	}

	if position == "" || len([]rune(position)) > 200 {
		return nil, domain.ErrInvalidInput
	}

	deptExists, err := u.deptRepo.Exists(deptID)
	if err != nil {
		return nil, err
	}
	if !deptExists {
		return nil, domain.ErrNotFound
	}

	emp := &domain.Employee{
		DepartmentID: deptID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAt,
	}

	if err := u.empRepo.Create(emp); err != nil {
		return nil, err
	}

	return emp, nil
}
