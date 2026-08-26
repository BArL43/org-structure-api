package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidInput  = errors.New("invalid input data")
	ErrAlreadyExists = errors.New("department with this name already exists under this parent")
	ErrCycleDetected = errors.New("cannot move department inside itself or its subtree")
	ErrRequiredField = errors.New("missing required field")
)

type Department struct {
	ID        int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string       `json:"name" gorm:"type:varchar(200);not null"`
	ParentID  *int64       `json:"parent_id" gorm:"index;default:null"`
	CreatedAt time.Time    `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	Employees []Employee   `json:"employees,omitempty" gorm:"foreignKey:DepartmentID"`
	Children  []Department `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

type DepartmentUpdate struct {
	Name        *string
	ParentID    *int64
	ParentIDSet bool
}

type DepartmentRepository interface {
	Create(ctx context.Context, dept *Department) error
	GetByID(ctx context.Context, id int64, depth int, includeEmployees bool) (*Department, error)
	Update(ctx context.Context, id int64, update DepartmentUpdate) error
	DeleteAndReassign(ctx context.Context, id int64, reassignToID int64) error
	DeleteCascade(ctx context.Context, id int64) error
	Exists(ctx context.Context, id int64) (bool, error)
	HasChildWithSameName(ctx context.Context, parentID *int64, name string, excludeID *int64) (bool, error)
	IsAncestor(ctx context.Context, ancestorID int64, descendantID int64) (bool, error)
}

type DepartmentUseCase interface {
	Create(ctx context.Context, name string, parentID *int64) (*Department, error)
	GetByID(ctx context.Context, id int64, depth int, includeEmployees bool) (*Department, error)
	Update(ctx context.Context, id int64, update DepartmentUpdate) (*Department, error)
	Delete(ctx context.Context, id int64, mode string, reassignID *int64) error
}
