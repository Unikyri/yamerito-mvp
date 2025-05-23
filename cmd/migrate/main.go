package main

import (
	"log"
	// "os" // Para argumentos de línea de comandos en el futuro

	"github.com/Unikyri/yamerito-mvp/internal/auth" // Importar para hashear contraseña
	"github.com/Unikyri/yamerito-mvp/internal/config"
	"github.com/Unikyri/yamerito-mvp/internal/database"
	"github.com/Unikyri/yamerito-mvp/internal/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func main() {
	log.Println("Iniciando proceso de migración...")

	// 1. Cargar configuración de la aplicación
	appConfig := config.LoadConfig()
	if appConfig == nil {
		log.Fatal("Error al cargar la configuración de la aplicación.")
	}
	log.Println("Configuración cargada.")

	// 2. Conectar a la base de datos
	// La función ConnectDB ya maneja el logging interno
	database.ConnectDB(appConfig)
	db := database.GetDB()
	if db == nil {
		log.Fatal("Error al obtener la instancia de la base de datos.")
	}
	log.Println("Conexión a la base de datos establecida para migración.")

	// 3. Definir las migraciones
	// Usaremos un ID basado en timestamp para la primera migración
	// (ej. YYYYMMDDHHMMSS_descriptive_name)
	// Esto ayuda a mantener el orden si añades más migraciones.
	migrations := []*gormigrate.Migration{
		{
			ID: "20250508030000_create_users_table", // Timestamp aproximado actual
			Migrate: func(tx *gorm.DB) error {
				log.Println("Ejecutando migración: creando tabla 'users'...")
				// AutoMigrate creará la tabla, columnas, índices, etc.
				// basados en la estructura del modelo models.User
				err := tx.AutoMigrate(&models.User{})
				if err == nil {
					log.Println("Tabla 'users' creada/actualizada exitosamente.")
				}
				return err
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Ejecutando rollback: eliminando tabla 'users'...")
				err := tx.Migrator().DropTable(&models.User{})
				if err == nil {
					log.Println("Tabla 'users' eliminada exitosamente.")
				}
				return err
			},
		},
		{
			ID: "20250508161000_create_employee_details_table", // Nuevo ID para esta migración
			Migrate: func(tx *gorm.DB) error {
				log.Println("Ejecutando migración: creando tabla 'employee_details'...")
				err := tx.AutoMigrate(&models.EmployeeDetail{})
				if err == nil {
					log.Println("Tabla 'employee_details' creada/actualizada exitosamente.")
				}
				return err
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Ejecutando rollback: eliminando tabla 'employee_details'...")
				err := tx.Migrator().DropTable(&models.EmployeeDetail{})
				if err == nil {
					log.Println("Tabla 'employee_details' eliminada exitosamente.")
				}
				return err
			},
		},
		// --- Seed para el Usuario Administrador ---
		{
			ID: "20250508150000_seed_admin_user", // Timestamp aproximado actual para el seed
			Migrate: func(tx *gorm.DB) error {
				log.Println("Ejecutando seed: creando usuario administrador 'wolfang'...")

				adminUsername := "wolfang"
				adminPassword := "w01f4ng@"

				// Verificar si el usuario admin ya existe
				var existingAdmin models.User
				if err := tx.Where("username = ?", adminUsername).First(&existingAdmin).Error; err == nil {
					log.Printf("Usuario administrador '%s' ya existe. Saltando seed.", adminUsername)
					return nil // Ya existe, no hacer nada
				} else if err != gorm.ErrRecordNotFound {
					// Otro error al buscar, retornar el error
					log.Printf("Error al verificar existencia del usuario admin '%s': %v", adminUsername, err)
					return err
				}

				// Hashear la contraseña del admin
				hashedPassword, err := auth.HashPassword(adminPassword, nil) // Usar nil para params por defecto
				if err != nil {
					log.Printf("Error al hashear la contraseña para el usuario admin '%s': %v", adminUsername, err)
					return err
				}

				adminUser := models.User{
					Username:     adminUsername,
					PasswordHash: hashedPassword,
					Role:         models.RoleAdmin,
				}

				if err := tx.Create(&adminUser).Error; err != nil {
					log.Printf("Error al crear usuario administrador '%s': %v", adminUsername, err)
					return err
				}

				log.Printf("Usuario administrador '%s' creado exitosamente.", adminUsername)
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				// En el rollback, podríamos optar por eliminar el usuario admin si se desea.
				// Sin embargo, para un seed, a veces el rollback se deja vacío o simplemente loguea.
				// Por seguridad y para no perder el admin accidentalmente, solo loguearemos.
				log.Println("Ejecutando rollback de seed_admin_user: No se eliminará el usuario admin 'wolfang' automáticamente.")
				// Si quisieras eliminarlo: 
				// return tx.Where("username = ? AND role = ?", "wolfang", models.RoleAdmin).Delete(&models.User{}).Error
				return nil
			},
		},
		// --- Fin del Seed para el Usuario Administrador ---
		// --- Aquí puedes añadir más migraciones en el futuro ---
		// {
		// 	ID: "YYYYMMDDHHMMSS_add_new_field_to_users",
		// 	Migrate: func(tx *gorm.DB) error {
		// 		return tx.Exec("ALTER TABLE users ADD COLUMN new_field VARCHAR(255);").Error
		//  },
		//  Rollback: func(tx *gorm.DB) error {
		// 		return tx.Exec("ALTER TABLE users DROP COLUMN new_field;").Error
		//  },
		// },
	}

	// 4. Inicializar Gormigrate
	// DefaultOptions incluye una tabla `migrations` para rastrear las migraciones aplicadas.
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)

	// 5. Ejecutar las migraciones
	// TODO: Añadir manejo de argumentos de línea de comandos (ej. "up", "down", "status")
	// Por ahora, simplemente aplicará todas las migraciones pendientes.
	log.Println("Aplicando migraciones pendientes...")
	if err := m.Migrate(); err != nil {
		log.Fatalf("Error durante la migración: %v", err)
	}

	log.Println("Proceso de migración completado exitosamente.")
}
