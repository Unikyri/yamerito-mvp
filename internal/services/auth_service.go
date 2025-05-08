package services

import (
	"errors"
	"log"

	"github.com/Unikyri/yamerito-mvp/internal/auth"
	"github.com/Unikyri/yamerito-mvp/internal/models"
	"gorm.io/gorm"
)

// LoginRequestDTO define la estructura para las solicitudes de login.
// Movida aquí ya que es específica de la autenticación.
type LoginRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthServiceInterface define la interfaz para operaciones de autenticación.
type AuthServiceInterface interface {
	LoginUser(dto LoginRequestDTO) (string, *models.User, error)
}

// AuthService implementa AuthServiceInterface.
type AuthService struct {
	DB *gorm.DB
}

// NewAuthService crea una nueva instancia de AuthService.
func NewAuthService(db *gorm.DB) AuthServiceInterface {
	return &AuthService{DB: db}
}

// LoginUser autentica a un usuario y devuelve un token JWT y los detalles del usuario.
func (s *AuthService) LoginUser(dto LoginRequestDTO) (string, *models.User, error) {
	var user models.User
	// Buscar usuario por nombre de usuario
	if err := s.DB.Where("username = ?", dto.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Intento de login para usuario no encontrado: %s", dto.Username)
			return "", nil, errors.New("usuario o contraseña incorrectos")
		}
		log.Printf("Error al buscar usuario '%s' durante login: %v", dto.Username, err)
		return "", nil, errors.New("error interno al intentar login")
	}

	// Verificar contraseña
	passwordMatch, err := auth.CheckPasswordHash(dto.Password, user.PasswordHash)
	if err != nil {
		log.Printf("Error al verificar hash de contraseña para usuario '%s': %v", dto.Username, err)
		return "", nil, errors.New("error interno al procesar login")
	}
	if !passwordMatch {
		log.Printf("Contraseña incorrecta para usuario: %s", dto.Username)
		return "", nil, errors.New("usuario o contraseña incorrectos")
	}

	// Generar token JWT
	// La firma es: GenerateJWT(userID uint, username string, role models.Role)
	token, err := auth.GenerateJWT(user.ID, user.Username, user.Role) // Corregido el orden de los argumentos
	if err != nil {
		log.Printf("Error al generar token JWT para usuario '%s': %v", dto.Username, err)
		return "", nil, errors.New("error al generar token de sesión")
	}

	log.Printf("Usuario '%s' logueado exitosamente.", user.Username)
	// No devolver la contraseña hasheada
	user.PasswordHash = ""
	return token, &user, nil
}
