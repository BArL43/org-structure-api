package repository

import (
	"context"
	"errors"

	"org-structure-api/internal/domain"

	"gorm.io/gorm"
)

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) domain.DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(ctx context.Context, dept *domain.Department) error {
	return translatePostgresError(r.db.WithContext(ctx).Create(dept).Error)
}

func (r *departmentRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Department{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *departmentRepository) HasChildWithSameName(ctx context.Context, parentID *int64, name string, excludeID *int64) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Department{}).Where("name = ?", name)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *departmentRepository) Update(ctx context.Context, id int64, update domain.DepartmentUpdate) error {
	updates := make(map[string]any, 2)
	if update.Name != nil {
		updates["name"] = *update.Name
	}
	if update.ParentIDSet {
		if update.ParentID == nil {
			updates["parent_id"] = nil
		} else {
			updates["parent_id"] = *update.ParentID
		}
	}
	if len(updates) == 0 {
		return nil
	}
	result := r.db.WithContext(ctx).Model(&domain.Department{}).Where("id = ?", id).Updates(updates)
	if err := translatePostgresError(result.Error); err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *departmentRepository) IsAncestor(ctx context.Context, ancestorID int64, descendantID int64) (bool, error) {
	const query = `
		WITH RECURSIVE sub_departments AS (
			SELECT id FROM departments WHERE id = ?
			UNION ALL
			SELECT d.id
			FROM departments d
			JOIN sub_departments sd ON d.parent_id = sd.id
		)
		SELECT EXISTS (
			SELECT 1 FROM sub_departments WHERE id = ?
		)
	`
	var exists bool
	if err := r.db.WithContext(ctx).Raw(query, ancestorID, descendantID).Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

func (r *departmentRepository) GetByID(ctx context.Context, id int64, depth int, includeEmployees bool) (*domain.Department, error) {
	var dept domain.Department
	query := r.db.WithContext(ctx).Model(&domain.Department{})

	if includeEmployees {
		query = query.Preload("Employees", func(db *gorm.DB) *gorm.DB {
			return db.Order("full_name ASC")
		})
	}

	currentPath := "Children"
	for level := 0; level < depth; level++ {
		query = query.Preload(currentPath)
		if includeEmployees {
			query = query.Preload(currentPath+".Employees", func(db *gorm.DB) *gorm.DB {
				return db.Order("full_name ASC")
			})
		}
		currentPath += ".Children"
	}

	err := query.First(&dept, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *departmentRepository) DeleteCascade(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&domain.Department{}, id)
	if err := translatePostgresError(result.Error); err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *departmentRepository) DeleteAndReassign(ctx context.Context, id int64, reassignToID int64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Employee{}).
			Where("department_id = ?", id).
			Update("department_id", reassignToID).Error; err != nil {
			return err
		}

		var currentDept domain.Department
		if err := tx.Select("parent_id").First(&currentDept, id).Error; err != nil {
			return err
		}

		var newParent any
		if currentDept.ParentID != nil {
			newParent = *currentDept.ParentID
		}
		if err := tx.Model(&domain.Department{}).
			Where("parent_id = ?", id).
			Update("parent_id", newParent).Error; err != nil {
			return err
		}

		return tx.Delete(&domain.Department{}, id).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	return translatePostgresError(err)
}
