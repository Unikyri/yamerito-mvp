package handlers

import (
	"log"
	"net/http"

	"github.com/Unikyri/yamerito-mvp/internal/models"
	"github.com/Unikyri/yamerito-mvp/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler maneja las solicitudes HTTP relacionadas con la autenticación.
type AuthHandler struct {
	AuthService services.AuthServiceInterface
}

// NewAuthHandler crea una nueva instancia de AuthHandler.
func NewAuthHandler(authService services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

// LoginResponseDTO define la estructura de la respuesta para un login exitoso.
// Incluye el token JWT y los detalles básicos del usuario.
// Movido aquí ya que es específico de la respuesta de autenticación.
type LoginResponseDTO struct {
	Token   string                `json:"token"`
	User    models.UserDetailDTO  `json:"user"` // Usamos UserDetailDTO para consistencia
	Message string                `json:"message"`
}

// LoginUser maneja las solicitudes de inicio de sesión.
func (h *AuthHandler) LoginUser(c *gin.Context) {
	var dto services.LoginRequestDTO // Usar el DTO del paquete services

	// Vincular y validar el JSON de la solicitud
	if err := c.ShouldBindJSON(&dto); err != nil {
		log.Printf("Error al vincular JSON para login: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Solicitud inválida", "details": err.Error()})
		return
	}

	// Llamar al servicio de login
	token, user, err := h.AuthService.LoginUser(dto)
	if err != nil {
		// El servicio ya debería loguear errores internos.
		if err.Error() == "usuario o contraseña incorrectos" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		}
		return
	}

	// Login exitoso
	c.JSON(http.StatusOK, LoginResponseDTO{
		Token: token,
		User: models.UserDetailDTO{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role, // Corregido: user.Role es models.Role, y el campo espera models.Role
			// EmployeeDetails no se incluyen en la respuesta de login por simplicidad,
			// pero podrían agregarse si es necesario obteniéndolos del 'user' retornado por AuthService.
		},
		Message: "Login exitoso",
	})
}

// RegisterAuthRoutes registra las rutas relacionadas con la autenticación.
func (h *AuthHandler) RegisterAuthRoutes(rg *gin.RouterGroup) {
	authRoutes := rg.Group("/auth") // Podríamos usar un prefijo /auth o directamente /users/login
	{
		authRoutes.POST("/login", h.LoginUser)
	}
}
