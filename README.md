# Yamerito MVP - Proyecto Capacitron

MVP "Yamerito" del proyecto Capacitron, construido con Wails, Go y React.

## Configuración del Proyecto Wails

Edita `wails.json` para configuraciones específicas. Más info: https://wails.io/docs/reference/project-config

## Requisitos Previos

*   Go (v1.24.1+)
*   Node.js (v18+ con npm)
*   Wails CLI (Instalar con `go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
*   Docker y Docker Compose (para la base de datos local)

## Configuración del Entorno Local

1.  **Clona el repositorio.**
2.  **Crea un archivo `.env`** en la raíz del proyecto (puedes copiar `env.example` si existe o crearlo desde cero):
    ```env
    DB_USER="yamerito_user"         # Usuario para Docker Compose MySQL
    DB_PASSWORD="your_strong_user_password" # Contraseña para Docker Compose MySQL
    DB_HOST="localhost"
    DB_PORT="3306"                  # Puerto de Docker Compose MySQL
    DB_NAME="yamerito_mvp_db"       # Nombre de BD para Docker Compose MySQL
    DB_SSL_MODE="false"             # O "REQUIRED" si configuras SSL para la BD local
    DB_CA_CERT_PATH=""            # Dejar vacío o ruta al cert si DB_SSL_MODE es true
    API_PORT="8080"
    JWT_SECRET_KEY="una_clave_secreta_muy_segura_para_desarrollo"
    GIN_MODE="debug"
    ```
    *Nota: Para conectar a tu base de datos de DigitalOcean localmente, ajusta las variables `DB_*` en `.env` a tus credenciales de DO, incluyendo `DB_SSL_MODE="REQUIRED"` y `DB_CA_CERT_PATH="./ca-certificate.crt"` (asegúrate de tener el archivo `ca-certificate.crt` en la raíz del proyecto).* 

## Ejecución Local

1.  **Iniciar Base de Datos (MySQL con Docker Compose):**
    Asegúrate de que `docker-compose.yml` esté configurado (especialmente las contraseñas).
    ```bash
    docker-compose up -d
    ```

2.  **Ejecutar Migraciones de Base de Datos (una sola vez o cuando cambien los modelos):**
    Esto creará las tablas necesarias. Asegúrate de que tu `.env` apunte a la base de datos correcta (local o remota).
    ```bash
    go run ./cmd/migrate/main.go
    ```

3.  **Iniciar el Backend (Servidor Go API):**
    Desde la raíz del proyecto:
    ```bash
    go run ./cmd/server/main.go
    ```
    El servidor API se ejecutará en `http://localhost:8080` (o el puerto definido en `API_PORT`).

4.  **Iniciar el Frontend (Wails con React en modo desarrollo):**
    Desde la raíz del proyecto, en otra terminal:
    ```bash
    wails dev
    ```
    Esto abrirá la aplicación de escritorio Wails. La interfaz de usuario también será accesible en un navegador (generalmente `http://localhost:3000` o similar, revisa la salida de `wails dev`) y las llamadas API se redirigirán al backend Go en el puerto `8080` gracias a la configuración de proxy en `frontend/vite.config.js`.

## Ejemplo de Inicio de Sesión

Una vez que el backend y el frontend estén corriendo:

1.  La aplicación mostrará la página de login.
2.  Ingresa las credenciales de un usuario existente. Por ejemplo, si creaste un usuario `testuser` con contraseña `password123` durante las pruebas o mediante el script de migración/seeding:
    *   **Usuario:** `testuser`
    *   **Contraseña:** `password123`
3.  Haz clic en "Login".
4.  Si las credenciales son correctas, serás redirigido o verás un mensaje de éxito, y un token JWT se almacenará en el `localStorage` del navegador (visible en las herramientas de desarrollo del frontend Wails o del navegador web).

## Building (Compilación para Producción)

Para construir un paquete redistribuible en modo producción:

```bash
wails build
```
Esto generará un ejecutable en el directorio `build/bin` que incluye tanto el backend Go como el frontend React compilado.
