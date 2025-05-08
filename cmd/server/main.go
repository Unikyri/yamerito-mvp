package main

import (
	"log"
	"net/http"

	"github.com/Unikyri/yamerito-mvp/internal/config"
	"github.com/Unikyri/yamerito-mvp/internal/database"
	"github.com/Unikyri/yamerito-mvp/internal/handlers"
	"github.com/Unikyri/yamerito-mvp/internal/auth"
	"github.com/Unikyri/yamerito-mvp/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Iniciando servidor API Yamerito MVP...")

	// Cargar configuración
	appConfig := config.LoadConfig()
	if appConfig == nil {
		log.Fatal("Error: No se pudo cargar la configuración.")
	}
	log.Println("Configuración cargada.")

	// Inicializar JWT Secret
	if err := auth.InitJWT(); err != nil { 
		log.Fatalf("Error al inicializar JWT: %v", err)
	}
	log.Println("Clave secreta JWT cargada.")

	// Conectar a la base de datos
	database.ConnectDB(appConfig)
	db := database.GetDB()
	if db == nil {
		log.Fatal("Error: No se pudo obtener la instancia de la base de datos.")
	}
	log.Println("Conexión a la base de datos establecida y obtenida para el servidor.")

	// Inicializar el router Gin
	// gin.SetMode(gin.ReleaseMode) // Descomentar para producción
	router := gin.Default() // Default() incluye logger y recovery middleware

	// Configurar CORS
	// Para desarrollo, podemos ser un poco más permisivos.
	// Para producción, deberías restringir los orígenes a tu dominio de frontend real.
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"} // Puerto común de Vite en `wails dev`
	// Si usas `wails build` y sirves desde file:// o un localhost diferente para el frontend en prod,
	// podrías necesitar añadir más orígenes o usar corsConfig.AllowAllOrigins = true (menos seguro).
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"} // Añadir Authorization para JWT
	router.Use(cors.New(corsConfig))

	// Rutas de prueba
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "API Yamerito MVP está funcionando!",
		})
	})

	// --- Configurar Handlers y Rutas de la API ---
	userHandler := handlers.NewUserHandler(db)

	// Agrupar rutas de la API bajo /api/v1
	v1 := router.Group("/api/v1")
	{
		userHandler.RegisterUserRoutes(v1) // Rutas públicas: /users/register, /users/login

		// Grupo de rutas autenticadas
		authRequired := v1.Group("") // Podría ser /auth o directamente bajo v1
		authRequired.Use(middleware.AuthMiddleware()) // Aplicar middleware JWT a este grupo
		{
			// Endpoint de ejemplo para obtener información del usuario autenticado
			authRequired.GET("/me", func(c *gin.Context) {
				claims, exists := middleware.GetAuthClaims(c)
				if !exists {
					// Esto no debería ocurrir si AuthMiddleware funciona correctamente
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener claims de autenticación"})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message": "Información del usuario autenticado",
					"user_id": claims.UserID,
					"username": claims.Username,
					"role":    claims.Role,
					"expires_at": claims.ExpiresAt.Time,
				})
			})

			// Ejemplo de ruta protegida solo para ADMINS:
			// adminRoutes := authRequired.Group("/admin")
			// adminRoutes.Use(middleware.AuthorizeRole(models.RoleAdmin))
			// {
			// 	adminRoutes.GET("/some-admin-action", func(c *gin.Context) {
			// 		c.JSON(http.StatusOK, gin.H{"message": "Acción de administrador realizada"})
			// 	})
			// }
		}
	}
	// --- Fin Configurar Handlers y Rutas de la API ---

	// Iniciar el servidor
	serverPort := config.GetEnv("API_SERVER_PORT", "8080") // Puedes definir esto en .env
	log.Printf("Servidor escuchando en el puerto %s...\n", serverPort)
	if err := router.Run(":" + serverPort); err != nil {
		log.Fatalf("Error al iniciar el servidor Gin: %v", err)
	}
}
