package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/Unikyri/yamerito-mvp/internal/auth"
	"github.com/Unikyri/yamerito-mvp/internal/models"
	"gorm.io/gorm"
)

// UserServiceInterface define la interfaz para los servicios de usuario.
type UserServiceInterface interface {
	//LoginUser(dto LoginRequestDTO) (string, *models.User, error)

	// Admin User Management
	CreateUserByAdmin(dto AdminCreateUserDTO) (*UserDetailDTO, error)
	ListUsers() ([]UserDetailDTO, error) // Devolver DTO para no exponer hash
	GetUserByID(id uint) (*UserDetailDTO, error)    // Devolver DTO
	UpdateUserByAdmin(id uint, dto AdminUpdateUserDTO) (*UserDetailDTO, error) // Devolver DTO
	DeleteUser(id uint) error
}

// UserService implementa UserServiceInterface.
type UserService struct {
	DB *gorm.DB
}

// NewUserService crea una nueva instancia de UserService.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

/* // Commenting out LoginRequestDTO as it's moved to auth_service.go
// LoginRequestDTO define la estructura para las solicitudes de login.
type LoginRequestDTO struct {
	Username string `json:"username" binding:"required,alphanum,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}
*/

// EmployeeDetailInputDTO define los datos para crear/actualizar los detalles de un empleado.
// Usado dentro de AdminCreateUserDTO y AdminUpdateUserDTO.
// Los campos son punteros para permitir actualizaciones parciales y distinguir entre no enviado y valor vacío.
// Sin embargo, para la creación, algunos podrían ser requeridos si la lógica de negocio lo impone.
// Por simplicidad inicial, los haremos opcionales en el DTO y el servicio validará según sea necesario.
// Si un campo es string, y se quiere que sea requerido, se puede quitar el puntero y añadir `binding:"required"`.
// Para este ejemplo, el Nombre y Apellido serán requeridos en la creación implícitamente por el servicio.
type EmployeeDetailInputDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	LastName    *string `json:"last_name" binding:"omitempty,min=1,max=100"`
	Email       *string `json:"email" binding:"omitempty,email,max=100"`
	PhoneNumber *string `json:"phone_number,omitempty" binding:"omitempty,max=20"`
	Position    *string `json:"position,omitempty" binding:"omitempty,max=100"`
}

// AdminCreateUserDTO define la estructura para que un administrador cree un nuevo usuario.
// Incluye detalles básicos del usuario y opcionalmente detalles del empleado.
type AdminCreateUserDTO struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Role     string `json:"role" binding:"omitempty,oneof=Admin Employee"` // Default a Employee si está vacío, y validación a PascalCase
	EmployeeDetails *EmployeeDetailInputDTO `json:"employee_details,omitempty"`
}

// AdminUpdateUserDTO define la estructura para que un admin actualice un usuario.
// Todos los campos son opcionales (punteros).
type AdminUpdateUserDTO struct {
	Username *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"` // Puntero para distinguir entre no enviado y vacío
	Password *string `json:"password,omitempty" binding:"omitempty,min=8,max=72"` // Puntero para cambio opcional
	Role     *string `json:"role,omitempty" binding:"omitempty,oneof=Admin Employee"` // Puntero, validación a PascalCase
	EmployeeDetails *EmployeeDetailInputDTO `json:"employee_details,omitempty"` // Para actualizar detalles del empleado
}

// UserDetailDTO define la estructura de datos detallada de un usuario para respuestas de API.
// Esta es la estructura que se devuelve en la mayoría de los endpoints que retornan información de usuario.
type UserDetailDTO struct {
	ID       uint                `json:"id"`
	Username string              `json:"username"`
	Role     string              `json:"role"`
	EmployeeDetails *models.EmployeeDetail `json:"employee_details,omitempty"` // Mostrar detalles del empleado
}

// LoginUser maneja la lógica de inicio de sesión de un usuario.
// Devuelve el token JWT, el objeto User y un error si ocurre alguno.
/*func (s *UserService) LoginUser(dto LoginRequestDTO) (string, *models.User, error) {
	// 1. Validar DTO (Gin ya lo hace a nivel de handler con `ShouldBindJSON`)

	// 2. Buscar usuario por Username
	var user models.User
	if err := s.DB.Where("username = ?", dto.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("usuario no encontrado o contraseña incorrecta") // Error genérico
		}
		log.Printf("Error al buscar usuario %s: %v", dto.Username, err)
		return "", nil, errors.New("error al intentar iniciar sesión")
	}

	// 3. Verificar contraseña
	match, err := auth.CheckPasswordHash(dto.Password, user.PasswordHash)
	if err != nil {
		log.Printf("Error al verificar hash de contraseña para %s: %v", dto.Username, err)
		return "", nil, errors.New("error al intentar iniciar sesión")
	}
	if !match {
		return "", nil, errors.New("usuario no encontrado o contraseña incorrecta") // Error genérico
	}

	// 4. Generar token JWT
	// user.Role ya es del tipo models.Role, así que se puede pasar directamente a GenerateJWT.
	tokenString, err := auth.GenerateJWT(user.ID, user.Username, user.Role) // <--- PASAR user.Role DIRECTAMENTE
	if err != nil {
		log.Printf("Error al generar token JWT para %s: %v", dto.Username, err)
		return "", nil, errors.New("error al intentar iniciar sesión")
	}

	// 5. Devolver el token y el usuario (sin el PasswordHash para mayor seguridad, aunque el handler se encarga del DTO)
	// user.PasswordHash = "" // Esto es opcional aquí ya que el handler usará UserDetailDTO

	return tokenString, &user, nil
}*/

// CreateUserByAdmin crea un nuevo usuario con rol y detalles especificados por un administrador.
func (s *UserService) CreateUserByAdmin(dto AdminCreateUserDTO) (*UserDetailDTO, error) {
	userRole, err := models.ParseRole(dto.Role)
	if err != nil {
		if dto.Role == "" { // Si el rol está vacío en el DTO, asignamos EMPLOYEE por defecto
			userRole = models.RoleEmployee
		} else {
			return nil, fmt.Errorf("rol inválido: %s", dto.Role)
		}
	}

	hashedPassword, err := auth.HashPassword(dto.Password, nil) // Usar auth.HashPassword
	if err != nil {
		log.Printf("Error al hashear contraseña durante creación por admin: %v", err)
		return nil, errors.New("error interno al procesar la contraseña")
	}

	newUser := models.User{
		Username:     dto.Username,
		PasswordHash: hashedPassword,
		Role:         userRole,
	}

	// Asignar EmployeeDetail si se proporcionan los datos
	if dto.EmployeeDetails != nil {
		empDetail := models.EmployeeDetail{}
		if dto.EmployeeDetails.Name != nil {
			empDetail.Name = *dto.EmployeeDetails.Name
		}
		if dto.EmployeeDetails.LastName != nil {
			empDetail.LastName = *dto.EmployeeDetails.LastName
		}
		if dto.EmployeeDetails.Email != nil {
			empDetail.Email = *dto.EmployeeDetails.Email
		}
		if dto.EmployeeDetails.PhoneNumber != nil {
			empDetail.PhoneNumber = *dto.EmployeeDetails.PhoneNumber
		}
		if dto.EmployeeDetails.Position != nil {
			empDetail.Position = *dto.EmployeeDetails.Position
		}
		newUser.EmployeeDetail = empDetail // Asignar al campo singular 'EmployeeDetail'
	}

	// Usar una transacción para asegurar que User y EmployeeDetail se creen atómicamente
	tx := s.DB.Begin()
	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		log.Printf("Error al crear usuario por admin en DB: %v", err)
		return nil, errors.New("no se pudo crear el usuario")
	}

	tx.Commit()

	detailDTO := UserDetailDTO{
		ID:       newUser.ID,
		Username: newUser.Username,
		Role:     string(newUser.Role),
	}
	if newUser.EmployeeDetail.ID != 0 {
		detailDTO.EmployeeDetails = &newUser.EmployeeDetail
	}

	return &detailDTO, nil
}

// ListUsers recupera una lista de todos los usuarios.
func (s *UserService) ListUsers() ([]UserDetailDTO, error) {
	var users []models.User
	if err := s.DB.Preload("EmployeeDetail").Find(&users).Error; err != nil {
		log.Printf("Error al listar usuarios: %v", err)
		return nil, errors.New("no se pudo obtener la lista de usuarios")
	}

	userDTOs := make([]UserDetailDTO, 0, len(users))
	for _, u := range users {
		dto := UserDetailDTO{
			ID:       u.ID,
			Username: u.Username,
			Role:     string(u.Role),
		}
		if u.EmployeeDetail.ID != 0 { 
			dto.EmployeeDetails = &u.EmployeeDetail
		}
		userDTOs = append(userDTOs, dto)
	}
	return userDTOs, nil
}

// GetUserByID recupera un usuario por su ID.
func (s *UserService) GetUserByID(id uint) (*UserDetailDTO, error) {
	var user models.User
	if err := s.DB.Preload("EmployeeDetail").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuario no encontrado")
		}
		log.Printf("Error al obtener usuario por ID (%d): %v", id, err)
		return nil, errors.New("no se pudo obtener el usuario")
	}

	detailDTO := UserDetailDTO{
		ID:       user.ID,
		Username: user.Username,
		Role:     string(user.Role),
	}
	if user.EmployeeDetail.ID != 0 { 
		detailDTO.EmployeeDetails = &user.EmployeeDetail
	}

	return &detailDTO, nil
}

// UpdateUserByAdmin actualiza los datos de un usuario existente.
func (s *UserService) UpdateUserByAdmin(id uint, dto AdminUpdateUserDTO) (*UserDetailDTO, error) {
	var user models.User
	tx := s.DB.Begin()

	if err := tx.Preload("EmployeeDetail").First(&user, id).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuario no encontrado para actualizar")
		}
		log.Printf("Error al buscar usuario %d para actualizar: %v", id, err)
		return nil, errors.New("error al buscar usuario")
	}

	updated := false

	if dto.Username != nil && *dto.Username != "" && *dto.Username != user.Username {
		user.Username = *dto.Username
		updated = true
	}
	if dto.Password != nil && *dto.Password != "" {
		hashedPassword, err := auth.HashPassword(*dto.Password, nil)
		if err != nil {
			tx.Rollback()
			log.Printf("Error al hashear contraseña durante actualización por admin: %v", err)
			return nil, errors.New("error interno al procesar la contraseña")
		}
		user.PasswordHash = hashedPassword
		updated = true
	}
	if dto.Role != nil && *dto.Role != "" {
		newRole, err := models.ParseRole(*dto.Role)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("rol inválido para actualización: %s", *dto.Role)
		}
		if newRole != user.Role {
			user.Role = newRole
			updated = true
		}
	}

	if dto.EmployeeDetails != nil {
		if user.EmployeeDetail.ID == 0 {
			user.EmployeeDetail = models.EmployeeDetail{UserID: user.ID} 
		}

		if dto.EmployeeDetails.Name != nil {
			user.EmployeeDetail.Name = *dto.EmployeeDetails.Name
			updated = true
		}
		if dto.EmployeeDetails.LastName != nil {
			user.EmployeeDetail.LastName = *dto.EmployeeDetails.LastName
			updated = true
		}
		if dto.EmployeeDetails.Email != nil {
			user.EmployeeDetail.Email = *dto.EmployeeDetails.Email
			updated = true
		}
		if dto.EmployeeDetails.PhoneNumber != nil {
			user.EmployeeDetail.PhoneNumber = *dto.EmployeeDetails.PhoneNumber
			updated = true
		}
		if dto.EmployeeDetails.Position != nil {
			user.EmployeeDetail.Position = *dto.EmployeeDetails.Position
			updated = true
		}
	}

	if !updated {
		tx.Rollback()
		currentDetailDTO := UserDetailDTO{
			ID:       user.ID,
			Username: user.Username,
			Role:     string(user.Role),
		}
		if user.EmployeeDetail.ID != 0 {
			currentDetailDTO.EmployeeDetails = &user.EmployeeDetail
		}
		return &currentDetailDTO, nil
	}

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		log.Printf("Error al actualizar usuario %d en DB: %v", id, err)
		return nil, errors.New("no se pudo actualizar el usuario")
	}

	tx.Commit()

	finalDetailDTO := UserDetailDTO{
		ID:       user.ID,
		Username: user.Username,
		Role:     string(user.Role),
	}
	if user.EmployeeDetail.ID != 0 {
		finalDetailDTO.EmployeeDetails = &user.EmployeeDetail
	}
	return &finalDetailDTO, nil
}

// DeleteUser elimina un usuario por su ID (borrado lógico si DeletedAt está configurado).
func (s *UserService) DeleteUser(id uint) error {
	result := s.DB.Delete(&models.User{}, id)
	if result.Error != nil {
		log.Printf("Error al eliminar usuario %d: %v", id, result.Error)
		return errors.New("error al eliminar el usuario")
	}
	if result.RowsAffected == 0 {
		return errors.New("usuario no encontrado para eliminar")
	}
	return nil
}
