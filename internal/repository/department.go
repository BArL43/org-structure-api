package repository

import (
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

func (r *departmentRepository) Create(dept *domain.Department) error {
	return r.db.Create(dept).Error
}

func (r *departmentRepository) Exists(id int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Department{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *departmentRepository) HasChildWithSameName(parentID *int64, name string) (bool, error) {
	var count int64
	query := r.db.Model(&domain.Department{}).Where("name = ?", name)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *departmentRepository) Update(dept *domain.Department) error {
	return r.db.Model(dept).Select("name", "parent_id").Updates(dept).Error
}

func (r *departmentRepository) IsAncesstor(parentID int64, childID int64) (bool, error) {
	query := `
		WITH RECURSIVE subdepartments AS (
			SELECT id FROM departmets WHERE id = ?
			UNION
			SELECT d.id FROM departments d
			INNER JOIN sub_departments sd ON d.parent_id = sd.id
		)
		SELECT EXISTS(SELECT 1 FROM sub_departments WHERE id = ?)
	`
	var exists bool
	err := r.db.Raw(query, parentID, childID).Scan(&exists).Error
	return exists, err
}

func (r *departmentRepository) GetByID(id int64, depth int, includeEmployees bool) (*domain.Department, error) {
	var dept domain.Department

	dbQuery := r.db.Model(&domain.Department{})

	if includeEmployees {
		dbQuery = dbQuery.Preload("Employees", func(db *gorm.DB) *gorm.DB {
			return db.Order("full_name ASC")
		})
	}
	currentPath := "Children"

	for i := 0; i < depth; i++ {
		dbQuery = dbQuery.Preload(currentPath)
		if includeEmployees {
			dbQuery = dbQuery.Preload(currentPath+".Employees", func(db *gorm.DB) *gorm.DB {
				return db.Order("full_name ASC")
			})
		}
		currentPath += ".Children"
	}

	dbQuery = dbQuery.Preload(currentPath)

	err := dbQuery.First(&dept, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &dept, nil
}

func (r *departmentRepository) DeleteCascade(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		queryIds := `
			WITH RECURSIVE sub_departments AS(
				DELECT id FROM departments WHERE id = ?
				UNION
				SELECT d.id FREOM departments d
				INNER JOID sub_departments sd ON .parent_id = sd.id
			)
			SELECT id From sub_departments;
		`
		var ids []int64
		if err := tx.Raw(queryIds, id).Scan(&ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		if err := tx.Where("id IN ?", ids).Delete(&domain.Department{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *departmentRepository) DeleteAndReassign(id int64, reassignToID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&domain.Employee{}).
			Where("department_id = ?", id).
			Update("department_id", reassignToID).Error
		if err != nil {
			return err
		}

		var currentDept domain.Department
		if err := tx.Select("parent_id").First(&currentDept, id).Error; err != nil {
			return err
		}

		err = tx.Model(&domain.Department{}).
			Where("parent_id = ?", id).
			Update("parent_id", currentDept.ParentID).Error
		if err != nil {
			return err
		}

		if err := tx.Delete(&domain.Department{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}
