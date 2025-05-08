package config

import (
	"log"
	"os"

	"github.com/joho/godotenv" // Para cargar .env
)

// DBConfig almacena la configuración de la base de datos
type DBConfig struct {
	Username     string
	Password     string
	Host         string
	Port         string
	DatabaseName string
	SSLMode      string
	SSLCertPath  string // Ruta al archivo ca-certificate.crt
	// DSN ya no es necesario aquí, se construye en database.go
}

// AppConfig almacena toda la configuración de la aplicación
type AppConfig struct {
	Database DBConfig
}

// LoadConfig carga la configuración de la aplicación desde variables de entorno
// y opcionalmente desde un archivo .env
func LoadConfig() *AppConfig {
	// Intentar cargar el archivo .env desde la raíz del proyecto.
	// Esto es útil para desarrollo local. En producción, las variables
	// de entorno deberían estar configuradas directamente en el sistema.
	// Obtener la ruta al directorio del ejecutable actual
	// exePath, err := os.Executable()
	// if err != nil {
	// 	log.Printf("Advertencia: No se pudo obtener la ruta del ejecutable: %v", err)
	// }
	// projectRoot := filepath.Dir(filepath.Dir(exePath)) // Asumiendo que el binario está en cmd/algo/bin
	// envPath := filepath.Join(projectRoot, ".env")

	// Forma más simple si ejecutas desde la raíz o cmd/
	// Para producción, las variables de entorno deben ser gestionadas por el sistema/orquestador.
	err := godotenv.Load() // Busca .env en el directorio actual y padres
	if err != nil {
		log.Println("Advertencia: No se pudo cargar el archivo .env. Se usarán variables de entorno del sistema si están disponibles.")
	}

	dbUser := GetEnv("DB_USER", "")
	dbPassword := GetEnv("DB_PASSWORD", "")
	dbHost := GetEnv("DB_HOST", "")
	dbPort := GetEnv("DB_PORT", "25060") // Default port si no se especifica
	dbName := GetEnv("DB_NAME", "")
	dbSSLMode := GetEnv("DB_SSL_MODE", "REQUIRED") // Default SSL mode
	dbSSLCertPath := GetEnv("DB_SSL_CERT_PATH", "") // Ruta al certificado CA

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbName == "" || dbSSLCertPath == "" {
		log.Fatal("Error: Faltan variables de entorno críticas para la base de datos (DB_USER, DB_PASSWORD, DB_HOST, DB_NAME, DB_SSL_CERT_PATH).")
	}

	return &AppConfig{
		Database: DBConfig{
			Username:     dbUser,
			Password:     dbPassword,
			Host:         dbHost,
			Port:         dbPort,
			DatabaseName: dbName,
			SSLMode:      dbSSLMode,
			SSLCertPath:  dbSSLCertPath,
		},
	}
}

// GetEnv recupera una variable de entorno o devuelve un valor por defecto.
// Si la variable es requerida y no se encuentra, podría ser mejor log.Fatal aquí
// o manejarlo en LoadConfig.
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if fallback == "" && (key == "DB_USER" || key == "DB_PASSWORD" || key == "DB_HOST" || key == "DB_NAME" || key == "DB_SSL_CERT_PATH") {
		// No hacer fatal aquí, LoadConfig lo chequeará
	}
	return fallback
}
