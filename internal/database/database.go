package database

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"github.com/Unikyri/yamerito-mvp/internal/config" // Ajusta esta ruta si es necesario
	mysqlDriver "github.com/go-sql-driver/mysql" // Renombrado para evitar colisión con el alias de gorm
	mysqlGorm "gorm.io/driver/mysql" // Driver GORM para MySQL, alias para claridad
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDB inicializa la conexión a la base de datos
func ConnectDB(cfg *config.AppConfig) {
	var err error

	dsn := ""
	dbSSLMode := cfg.Database.SSLMode

	if dbSSLMode == "REQUIRED" || dbSSLMode == "VERIFY_CA" || dbSSLMode == "VERIFY_FULL" {
		// Cargar el certificado CA
		rootCertPool := x509.NewCertPool()
		pem, errCert := os.ReadFile(cfg.Database.SSLCertPath)
		if errCert != nil {
			log.Fatalf("Error al leer el archivo del certificado CA (%s): %v", cfg.Database.SSLCertPath, errCert)
		}
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			log.Fatal("Error al añadir el certificado CA al pool")
		}

		// Registrar la configuración TLS personalizada
		tlsConfigName := "custom_tls_yamerito" // Usar un nombre único para evitar colisiones si se registran múltiples
		err = mysqlDriver.RegisterTLSConfig(tlsConfigName, &tls.Config{
			RootCAs:    rootCertPool,
			MinVersion: tls.VersionTLS12, // O la versión que requiera DigitalOcean
		})
		if err != nil {
			log.Fatalf("Error al registrar la configuración TLS '%s': %v", tlsConfigName, err)
		}

		// Construir DSN para SSL
		// Formato: "user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local&tls=custom_tls_yamerito"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DatabaseName,
			tlsConfigName, // Usar la configuración TLS registrada
		)
		log.Println("Conectando a la base de datos con SSL/TLS (certificado CA personalizado)...")
	} else if dbSSLMode == "PREFERRED" || dbSSLMode == "skip-verify" { // O "true" para SSL simple sin verificación de CA
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=true", // o skip-verify
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DatabaseName,
		)
		log.Println("Conectando a la base de datos con SSL/TLS (skip-verify o preferred)...")
	} else { // "false" o no SSL
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DatabaseName,
		)
		log.Println("Conectando a la base de datos sin SSL/TLS...")
	}

	// Abrir la conexión a la base de datos
	DB, err = gorm.Open(mysqlGorm.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // O logger.Silent en producción
	})

	if err != nil {
		log.Fatalf("Error al conectar con la base de datos: %v\nDSN: %s", err, dsn)
	}

	log.Println("Conexión a la base de datos establecida exitosamente.")

	// (Opcional) Configurar pool de conexiones
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Error al obtener el objeto DB subyacente: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	// sqlDB.SetConnMaxLifetime(time.Hour)
}

// GetDB retorna la instancia de la base de datos
func GetDB() *gorm.DB {
	return DB
}
