package handlers

import (
	"net/http"
	"strconv"

	"github.com/Unikyri/yamerito-mvp/internal/services"
	"github.com/gin-gonic/gin"
)

// UserHandler maneja las solicitudes HTTP relacionadas con los usuarios.
type UserHandler struct {
	UserService services.UserServiceInterface
}

// NewUserHandler crea una nueva instancia de UserHandler.
// Ahora acepta UserServiceInterface para una mejor inyección de dependencias.
func NewUserHandler(userService services.UserServiceInterface) *UserHandler {
	return &UserHandler{UserService: userService}
}

// RegisterUserRoutes registra las rutas relacionadas con usuarios en un grupo de rutas de Gin.
func (h *UserHandler) RegisterUserRoutes(rg *gin.RouterGroup) {
	// userRoutes := rg.Group("/users")
	// {
		// userRoutes.POST("/register", h.RegisterUser) // Ruta eliminada
		// userRoutes.POST("/login", h.LoginUser) // Login route removed
	// }
}

// --- Admin User Management Handlers ---

// CreateUserByAdmin maneja la creación de un nuevo usuario por un administrador.
// POST /api/v1/admin/users
func (h *UserHandler) CreateUserByAdmin(c *gin.Context) {
	var dto services.AdminCreateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Solicitud inválida", "details": err.Error()})
		return
	}

	user, err := h.UserService.CreateUserByAdmin(dto)
	if err != nil {
		if err.Error() == "el nombre de usuario ya está en uso" || err.Error() == "rol proporcionado inválido" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear el usuario"})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Usuario creado exitosamente por admin", "user": user})
}

// ListUsers maneja la solicitud para listar todos los usuarios.
// GET /api/v1/admin/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.UserService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al listar usuarios"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUserByID maneja la solicitud para obtener un usuario por su ID.
// GET /api/v1/admin/users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}

	user, err := h.UserService.GetUserByID(uint(id))
	if err != nil {
		if err.Error() == "usuario no encontrado" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el usuario"})
		}
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateUserByAdmin maneja la actualización de un usuario por un administrador.
// PUT /api/v1/admin/users/:id
func (h *UserHandler) UpdateUserByAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}

	var dto services.AdminUpdateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Solicitud inválida", "details": err.Error()})
		return
	}

	// No permitir que un admin se quite el rol de admin a sí mismo si es el único admin
	// (Esta lógica es compleja y podría ir en el servicio o requerir más contexto sobre cómo identificar al "yo")
	// Por ahora, se permite.

	user, err := h.UserService.UpdateUserByAdmin(uint(id), dto)
	if err != nil {
		if err.Error() == "usuario no encontrado para actualizar" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "el nuevo nombre de usuario ya está en uso por otro usuario" || err.Error() == "rol proporcionado inválido para la actualización" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar el usuario"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Usuario actualizado exitosamente", "user": user})
}

// DeleteUser maneja la eliminación de un usuario por un administrador.
// DELETE /api/v1/admin/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}

	// No permitir que un admin se elimine a sí mismo
	// Se necesitaría el ID del usuario autenticado para esta comprobación.
	// Esta lógica podría ir en el servicio.
	// userID := c.MustGet("userID").(uint) // Asumiendo que el middleware de auth lo añade
	// if uint(id) == userID {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "No puedes eliminarte a ti mismo"})
	// 	return
	// }

	err = h.UserService.DeleteUser(uint(id))
	if err != nil {
		if err.Error() == "usuario no encontrado para eliminar" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar el usuario"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Usuario eliminado exitosamente"})
}

// RegisterAdminUserRoutes registra las rutas CRUD para la gestión de usuarios por administradores.
func (h *UserHandler) RegisterAdminUserRoutes(rg *gin.RouterGroup) {
	adminUserRoutes := rg.Group("/users") // Corregido: Rutas bajo /api/v1/admin/users
	{
		adminUserRoutes.POST("", h.CreateUserByAdmin)
		adminUserRoutes.GET("", h.ListUsers)
		adminUserRoutes.GET("/:id", h.GetUserByID)
		adminUserRoutes.PUT("/:id", h.UpdateUserByAdmin)
		adminUserRoutes.DELETE("/:id", h.DeleteUser)
	}
}

// TODO: Implementar un middleware de autenticación JWT para proteger rutas.
// TODO: Implementar la función ValidateJWT en auth/jwt.go.
