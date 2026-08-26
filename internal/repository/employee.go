package repository

import (
	"context"

	"org-structure-api/internal/domain"

	"gorm.io/gorm"
)

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) domain.EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, emp *domain.Employee) error {
	return translatePostgresError(r.db.WithContext(ctx).Create(emp).Error)
}
