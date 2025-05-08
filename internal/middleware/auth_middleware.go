package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/Unikyri/yamerito-mvp/internal/auth" // Para ValidateJWT
	"github.com/Unikyri/yamerito-mvp/internal/models" // Para models.Role
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload" // Clave para guardar los claims en el contexto de Gin
)

// AuthMiddleware crea un middleware de Gin para la autenticación JWT.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeaderKey)
		if len(authHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "falta encabezado de autorización"})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "formato de encabezado de autorización inválido"})
			return
		}

		authType := strings.ToLower(fields[0])
		if authType != authorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "tipo de autorización no soportado: " + authType})
			return
		}

		accessToken := fields[1]
		claims, err := auth.ValidateJWT(accessToken)
		if err != nil {
			log.Printf("Error al validar token JWT: %v", err) // Loguear el error específico
			// Devolver un error genérico al cliente por seguridad
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido o expirado"})
			return
		}

		// Guardar los claims en el contexto de Gin para uso posterior en los handlers
		c.Set(authorizationPayloadKey, claims)
		c.Next() // Continuar con el siguiente handler en la cadena
	}
}

// AuthorizeRole es un middleware para verificar si el usuario tiene un rol específico.
// Debe usarse DESPUÉS de AuthMiddleware.
func AuthorizeRole(requiredRole models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, exists := c.Get(authorizationPayloadKey)
		if !exists {
			log.Println("Error: payload de autorización no encontrado en el contexto. Asegúrate de que AuthMiddleware se ejecute primero.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
			return
		}

		claims, ok := payload.(*auth.Claims)
		if !ok {
			log.Println("Error: el payload de autorización en el contexto tiene un tipo inesperado.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
			return
		}

		if claims.Role != requiredRole {
			log.Printf("Acceso denegado para el usuario %s (rol %s). Se requiere rol: %s", claims.Username, claims.Role, requiredRole)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no tienes permiso para realizar esta acción"})
			return
		}

		c.Next()
	}
}

// GetAuthClaims recupera los claims de autenticación del contexto de Gin.
// Es una función helper para los handlers.
func GetAuthClaims(c *gin.Context) (*auth.Claims, bool) {
	payload, exists := c.Get(authorizationPayloadKey)
	if !exists {
		return nil, false
	}
	claims, ok := payload.(*auth.Claims)
	return claims, ok
}
