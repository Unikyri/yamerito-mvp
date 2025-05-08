package handlers

import (
	"log"
	"net/http"

	"github.com/Unikyri/yamerito-mvp/internal/models"
	"github.com/Unikyri/yamerito-mvp/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler maneja las solicitudes HTTP relacionadas con los usuarios.
type UserHandler struct {
	UserService *services.UserService
}

// NewUserHandler crea una nueva instancia de UserHandler.
// Es común inyectar el servicio aquí.
func NewUserHandler(db *gorm.DB) *UserHandler {
	userService := services.NewUserService(db)
	return &UserHandler{UserService: userService}
}

// RegisterUser maneja la solicitud de registro de un nuevo usuario.
// POST /api/v1/users/register
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var dto services.RegisterUserDTO

	// 1. Bind JSON a DTO y validar
	// c.ShouldBindJSON se encarga de parsear el cuerpo de la solicitud JSON
	// y aplicar las validaciones definidas en las etiquetas 'binding' del DTO.
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Solicitud inválida", "details": err.Error()})
		return
	}

	// 2. Llamar al servicio para registrar al usuario
	user, err := h.UserService.RegisterUser(dto)
	if err != nil {
		// Determinar el código de estado HTTP apropiado basado en el error del servicio
		// Por ejemplo, si es "usuario ya existe", podría ser http.StatusConflict
		// Por ahora, usaremos http.StatusInternalServerError para errores generales del servicio
		// y http.StatusBadRequest para errores de validación que el servicio podría devolver explícitamente.
		if err.Error() == "el nombre de usuario ya está en uso" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if err.Error() == "el nombre de usuario y la contraseña son obligatorios" { // Ejemplo de error de validación
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al registrar el usuario"})
		}
		return
	}

	// 3. Devolver respuesta exitosa
	// Es buena práctica no devolver la contraseña, incluso hasheada.
	// El servicio ya se encarga de limpiar el campo Password.
	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado exitosamente", "user": user})
}

// LoginResponseDTO define la estructura de la respuesta para un login exitoso.
// Incluye el token JWT y los detalles básicos del usuario.
type LoginResponseDTO struct {
	Token    string             `json:"token"`
	User     models.UserDetailDTO `json:"user"`
	Message  string             `json:"message"`
}

// LoginUser maneja las solicitudes de inicio de sesión.
func (h *UserHandler) LoginUser(c *gin.Context) {
	var dto services.LoginRequestDTO // Usar el DTO del paquete services

	// Vincular y validar el JSON de la solicitud
	if err := c.ShouldBindJSON(&dto); err != nil {
		log.Printf("Error al vincular JSON para login: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Solicitud inválida", "details": err.Error()})
		return
	}

	// Llamar al servicio de login
	token, user, err := h.UserService.LoginUser(dto) // Corregido: userService -> UserService
	if err != nil {
		// El servicio ya debería loguear errores internos.
		// Aquí decidimos qué error devolver al cliente.
		// "usuario no encontrado o contraseña incorrecta" es un error común para ambos casos.
		if err.Error() == "usuario no encontrado o contraseña incorrecta" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		} else {
			// Para otros errores (ej. "error al intentar iniciar sesión" del servicio)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		}
		return
	}

	// Login exitoso
	// Devolver el token y la información básica del usuario (sin el hash de la contraseña)
	c.JSON(http.StatusOK, LoginResponseDTO{
		Token: token,
		User: models.UserDetailDTO{
			ID:       user.ID,
			Username: user.Username,
			Role:     models.Role(user.Role), // Asegurar conversión a models.Role
		},
		Message: "Login exitoso",
	})
}

// RegisterUserRoutes registra las rutas relacionadas con usuarios en un grupo de rutas de Gin.
func (h *UserHandler) RegisterUserRoutes(rg *gin.RouterGroup) {
	userRoutes := rg.Group("/users")
	{
		userRoutes.POST("/register", h.RegisterUser)
		userRoutes.POST("/login", h.LoginUser) // Añadida ruta de login
	}
}

// TODO: Considerar un DTO específico para la respuesta del usuario en RegisterUser
// para no exponer toda la estructura del modelo directamente (ej. PasswordHash vacío).

// TODO: Implementar un middleware de autenticación JWT para proteger rutas.
// TODO: Implementar la función ValidateJWT en auth/jwt.go.
