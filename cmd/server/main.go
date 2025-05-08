package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Unikyri/yamerito-mvp/internal/config"
	"github.com/Unikyri/yamerito-mvp/internal/database"
	"github.com/Unikyri/yamerito-mvp/internal/handlers"
	"github.com/Unikyri/yamerito-mvp/internal/auth"
	"github.com/Unikyri/yamerito-mvp/internal/middleware"
	"github.com/Unikyri/yamerito-mvp/internal/models"
	"github.com/Unikyri/yamerito-mvp/internal/services"

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
	// Inicializar servicios
	authSvc := services.NewAuthService(db)
	userSvc := services.NewUserService(db) // NewUserService devuelve *UserService, que implementa UserServiceInterface

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	userHandler := handlers.NewUserHandler(userSvc) 

	// Agrupar rutas de la API bajo /api/v1
	apiV1 := router.Group("/api/v1")
	{
		// Rutas de autenticación (login)
		authHandler.RegisterAuthRoutes(apiV1)

		// Rutas de usuario (login, etc. - las que queden públicas o semi-públicas)
		// userHandler.RegisterUserRoutes(apiV1) // Esta función ahora está vacía o eliminada, ya que el login se movió.

		// Rutas de administración para gestión de usuarios
		// Estas rutas requieren autenticación y rol de Admin.
		adminRoutes := apiV1.Group("/admin")
		adminRoutes.Use(middleware.AuthMiddleware())                       // Primero, autenticar JWT
		adminRoutes.Use(middleware.AuthorizeRole(models.RoleAdmin)) // Luego, verificar rol Admin
		{
			// Aquí registramos las rutas que userHandler expondrá para /admin/users/*
			userHandler.RegisterAdminUserRoutes(adminRoutes) // Pasamos el grupo adminRoutes
		}

		// Grupo de rutas autenticadas
		authRequired := apiV1.Group("") // Podría ser /auth o directamente bajo v1
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
		}
	}
	// --- Fin Configurar Handlers y Rutas de la API ---

	// --- Servir Frontend --- 
	// Obtener la ruta del ejecutable para construir rutas relativas de forma segura
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error al obtener la ruta del ejecutable: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	// Asumimos que 'frontend/dist' está al mismo nivel que el ejecutable o en una estructura conocida
	// En Docker, copiaremos 'frontend/dist' a './frontend/dist' relativo a donde esté el server.
	staticFilesPath := filepath.Join(exeDir, "frontend", "dist") 
	// Si en Docker el binario está en /app/yamerito-server y los assets en /app/frontend/dist, 
	// entonces la ruta relativa desde el binario (en /app) sería "./frontend/dist"
	// Para simplificar y hacerlo más robusto en Docker, usaremos una ruta relativa directa que esperamos
	// que esté presente donde se ejecute el binario.
	// Este path funcionará si el directorio 'frontend/dist' está al lado del ejecutable.
	// En el Dockerfile, nos aseguraremos de que esto sea así.
	router.Static("/assets", filepath.Join(staticFilesPath, "assets")) // Servir assets JS/CSS etc.

	// Servir index.html para la raíz y cualquier otra ruta no API (catch-all para SPA)
	router.NoRoute(func(c *gin.Context) {
		// Solo interceptar si no es una ruta de API para evitar conflictos
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:5] == "/api/" {
			c.Next() // Dejar que el router maneje la ruta API (que resultará en 404 si no existe)
			return
		}
		c.File(filepath.Join(staticFilesPath, "index.html"))
	})
	// --- Fin Servir Frontend ---

	// Iniciar el servidor
	serverPort := config.GetEnv("API_SERVER_PORT", "8080") // Puedes definir esto en .env
	log.Printf("Servidor escuchando en el puerto %s...\n", serverPort)
	if err := router.Run(":" + serverPort); err != nil {
		log.Fatalf("Error al iniciar el servidor Gin: %v", err)
	}
}
