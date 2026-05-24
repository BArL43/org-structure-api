package domain

import "time"

type Employee struct {
	ID int64 `json:"id" gorm:"primaryKey:autoIncrement:true"`
	DepartmentID int64 `json:"department_id" gorm:"not null;index"`
	FullName string `json:"full_name" gorm:"type:varchar(200);not null"`
	Position string `json:"position" gorm:"type:varchar(200);not null"`
	HiredAt *time.Time `json:"hired_at" gorm:"default:null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

type EmployeeRepository interface {
	Create(emp *Employee) error
}

type EmployeeUseCase interface {
	Create(deptID int64, fullName, position string, hiredAt *time.Time) (*Employee, error)
}
