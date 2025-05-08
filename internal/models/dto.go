package models

// UserDetailDTO representa los datos públicos de un usuario.
// Se utiliza para evitar exponer campos sensibles como el hash de la contraseña.
type UserDetailDTO struct {
	ID       uint   `json:"id"`       // Corregido json:"id"
	Username string `json:"username"` // Corregido json:"username"
	Role     Role   `json:"role"`     // Corregido json:"role"
}
