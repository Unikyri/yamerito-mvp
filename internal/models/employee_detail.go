package models

import (
	"time"

	"gorm.io/gorm"
)

// EmployeeDetail contiene información detallada sobre un empleado, vinculada a un User.
type EmployeeDetail struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	UserID        uint           `gorm:"uniqueIndex;not null" json:"user_id"` // Clave foránea a users.id, debe ser única
	Name          string         `gorm:"size:100" json:"name"`
	LastName      string         `gorm:"size:100" json:"last_name"`
	Email         string         `gorm:"size:100;uniqueIndex" json:"email"` // Email del empleado, también único
	PhoneNumber   string         `gorm:"size:20;index" json:"phone_number,omitempty"`
	Position      string         `gorm:"size:100" json:"position,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
