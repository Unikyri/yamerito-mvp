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
	RegisterUser(dto RegisterUserDTO) (*models.User, error)
	LoginUser(dto LoginRequestDTO) (string, *models.User, error)
}

// UserService implementa UserServiceInterface.
type UserService struct {
	DB *gorm.DB
}

// NewUserService crea una nueva instancia de UserService.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

// RegisterUserDTO (Data Transfer Object) para el registro de usuarios.
type RegisterUserDTO struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=100"` // Reglas de validación para Gin
	Role     string `json:"role,omitempty" binding:"omitempty,oneof=ADMIN EMPLOYEE"` // omitempty, valor por defecto "EMPLOYEE" se maneja en lógica
}

// LoginRequestDTO define la estructura para las solicitudes de login.
type LoginRequestDTO struct {
	Username string `json:"username" binding:"required,alphanum,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}

// RegisterUser maneja la creación de un nuevo usuario.
func (s *UserService) RegisterUser(dto RegisterUserDTO) (*models.User, error) {
	// 1. Validar entrada (Gin lo hace con 'binding', pero podemos añadir más aquí si es necesario)
	if dto.Username == "" || dto.Password == "" {
		return nil, errors.New("el nombre de usuario y la contraseña son obligatorios")
	}
	// Podríamos añadir validación de longitud/formato aquí también, aunque Gin ayuda.

	// 2. Verificar si el usuario ya existe
	var existingUser models.User
	err := s.DB.Where("username = ?", dto.Username).First(&existingUser).Error
	if err == nil { // Usuario encontrado
		return nil, errors.New("el nombre de usuario ya está en uso")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) { // Otro error de base de datos
		log.Printf("Error al verificar si el usuario existe: %v", err)
		return nil, fmt.Errorf("error al verificar la base de datos: %w", err)
	}

	// 3. Hashear la contraseña
	hashedPassword, err := auth.HashPassword(dto.Password, nil) // Usar DefaultParams
	if err != nil {
		log.Printf("Error al hashear la contraseña: %v", err)
		return nil, errors.New("error al procesar la contraseña")
	}

	// 4. Determinar el rol
	userRole := dto.Role
	if userRole == "" {
		userRole = "employee" // Valor por defecto
	}
	// Podríamos validar el rol aquí si tenemos una lista finita de roles permitidos

	// 5. Crear el objeto User
	newUser := models.User{
		Username:     dto.Username,
		PasswordHash: hashedPassword,
		Role:         models.Role(userRole),
	}

	// 6. Guardar el usuario en la base de datos
	result := s.DB.Create(&newUser)
	if result.Error != nil {
		log.Printf("Error al guardar el nuevo usuario: %v", result.Error)
		return nil, errors.New("no se pudo crear el usuario en la base de datos")
	}

	log.Printf("Usuario '%s' registrado exitosamente con ID %d", newUser.Username, newUser.ID)
	// Es buena práctica no devolver la contraseña hasheada, incluso si es un hash.
	// Para DTOs de respuesta, podrías crear uno que omita el campo Password.
	// Por ahora, devolvemos el modelo completo para simplicidad.
	newUser.PasswordHash = "" // Limpiar el hash de la respuesta para el cliente
	return &newUser, nil
}

// LoginUser maneja la lógica de inicio de sesión de un usuario.
// Devuelve el token JWT, el objeto User y un error si ocurre alguno.
func (s *UserService) LoginUser(dto LoginRequestDTO) (string, *models.User, error) {
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
}
