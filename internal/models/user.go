package models

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Role define los roles de usuario en el sistema
type Role string

const (
	RoleAdmin    Role = "ADMIN"
	RoleEmployee Role = "EMPLOYEE"
)

func (r Role) String() string {
	return string(r)
}

// ParseRole convierte una cadena a un tipo Role.
// Devuelve un error si la cadena no es un rol válido.
func ParseRole(s string) (Role, error) {
	s = strings.ToUpper(strings.TrimSpace(s)) // Normalizar: a mayúsculas y sin espacios extra
	switch s {
	case string(RoleAdmin):
		return RoleAdmin, nil
	case string(RoleEmployee):
		return RoleEmployee, nil
	default:
		return "", fmt.Errorf("rol inválido: '%s'", s)
	}
}

// User define el modelo de usuario para la base de datos
type User struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"` // Timestamp de creación (automático por GORM)
	UpdatedAt time.Time      `json:"updated_at"` // Timestamp de última actualización (automático por GORM)
	DeletedAt gorm.DeletedAt `gorm:"index"`      // Para borrado lógico (soft delete)

	Username     string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"` // No exponer en JSON por defecto
	Role         Role   `gorm:"type:varchar(20);not null" json:"role"`

	// Relación One-to-One con EmployeeDetail
	// El UserID en EmployeeDetail apuntará a este User.
	// Usamos SET NULL para OnDelete para que si se borra el usuario, el employee_detail.user_id se vuelva NULL,
	// pero el registro EmployeeDetail podría conservarse si se desea (o eliminarse por separado).
	// Si se quiere borrar en cascada el EmployeeDetail cuando se borra el User, usar OnDelete:CASCADE.
	EmployeeDetail EmployeeDetail `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"employee_details,omitempty"`

	// Podríamos añadir más campos aquí si son necesarios para el perfil,
	// como FirstName, LastName, Email, IsActive, etc.
	// Por ahora, nos centramos en lo esencial para RF-MVP1.
	// FirstName string `gorm:"size:100"`
	// LastName  string `gorm:"size:100"`
	// Email     string `gorm:"uniqueIndex;size:100"`
	// IsActive  bool   `gorm:"default:true"`
}

// Puedes añadir métodos al modelo User aquí si es necesario, por ejemplo,
// para validar la contraseña (aunque eso usualmente va en un paquete de servicio/handler).

/*
Consideraciones para RF-MVP1:
RF-MVP1.1: Crear cuentas (username, contraseña inicial) -> Cubierto por Username, PasswordHash. La "contraseña inicial" será procesada para generar el hash.
RF-MVP1.2: Editar info básica -> Se pueden añadir campos y luego permitir su edición.
RF-MVP1.3: Suspender/eliminar -> 'IsActive' o 'DeletedAt' (borrado lógico) pueden cubrir esto. GORM ya soporta borrado lógico con gorm.DeletedAt.
RF-MVP1.4: Iniciar sesión -> Implica verificar Username y PasswordHash.
RF-MVP1.5: Roles "Empleado" y "Administrador" -> Cubierto por el campo 'Role'.
*/
