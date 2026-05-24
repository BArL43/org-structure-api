package domain

import (
	"time"
	"errors"
)

var (
	ErrNotFound      = errors.New("recource not found")
	ErrInvalidInput  = errors.New("invalid input data")
	ErrAlreadyExists = errors.New("department with this name already exists under this parent")
	ErrCycleDetected = errors.New("cannot move department inside itself or its subtree")
	ErrRequiredField = errors.New("missing required field")
)

type Department struct {
	ID 			int64	  `json:"id" gorm:"primaryKey;autoIncrement:true"`
	Name		string    `json:"name" gorm:"type:varchar(200);not null"`
	ParentID    *int64     `json:"parent_id" gorm:"index;default:null"`
	CreatedAt   time.Time `json:"created_at" gorm:"notnull;default:CURRENT_TIMESTAMP"`

	Employees   []Employee `json:"employees,omitempty" gorm:"foreignKey:DepartmentID"`
	Children    []Department `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

type DepartmentRepository interface {
	Create(dept *Department) error
	GetByID(id int64, depth int, includeEmployees bool) (*Department, error)
	Update(dept *Department) error
	DeleteAndReassign(id int64, reassignToID int64) error
	DeleteCascade(id int64) error

	Exists(id int64) (bool, error)
	HasChildWithSameName(parentID *int64, name string) (bool, error)
	IsAncesstor(parentID int64, childID int64) (bool, error)
}

type DepartmentUseCase interface {
	Create(name string, parentID *int64) (*Department, error)
	GetByID(id int64, depth int, includeEmployees bool) (*Department, error)
	Update(id int64, name *string, parentID *int64) (*Department,  error)
	Delete(id int64, mode string, reassignID *int64) error
}
