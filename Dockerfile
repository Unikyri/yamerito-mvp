# Etapa 1: Build del Frontend y Wails CLI
FROM golang:1.24.1-alpine AS builder

# Instalar dependencias necesarias para Wails (Node.js, npm, librerías C)
# Alpine usa 'apk' como gestor de paquetes
RUN apk add --no-cache nodejs npm git gcc musl-dev

# Instalar Wails CLI
RUN go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Establecer el directorio de trabajo para el frontend
WORKDIR /app/frontend

# Copiar archivos de package.json y package-lock.json (o yarn.lock)
COPY frontend/package.json frontend/package-lock.json* ./

# Instalar dependencias del frontend
RUN npm install

# Copiar el resto del código del frontend
COPY frontend/ .

# Construir el frontend
RUN npm run build

# Establecer el directorio de trabajo para la raíz del proyecto Go
WORKDIR /app

# Copiar el código Go y archivos relacionados (go.mod, go.sum, etc.)
COPY go.mod go.sum ./
COPY cmd cmd/
COPY internal internal/
COPY app.go .
COPY wails.json ./
# Copiar cualquier otro archivo .go o directorio necesario en la raíz o subdirectorios
# Ejemplo: COPY main.go .

# El directorio 'frontend' con 'frontend/dist' ya está en /app/frontend desde los pasos anteriores.
# Wails build debería encontrarlo automáticamente.

# Construir la aplicación Wails. El binario se llamará 'yamerito-mvp' (por tu go.mod o wails.json)
# Usamos el nombre 'yamerito-server' para el output para claridad.
RUN wails build -trimpath -ldflags="-s -w" -o yamerito-server

# Etapa 2: Imagen Final (Runtime)
FROM alpine:latest AS final

# Instalar certificados CA para conexiones HTTPS si tu app los necesita
RUN apk --no-cache add ca-certificates

# Establecer el directorio de trabajo
WORKDIR /app

# Copiar el binario construido desde la etapa 'builder'
COPY --from=builder /app/yamerito-server .

# Copiar el certificado CA si lo necesitas en el contenedor de producción
# Asegúrate de que ca-certificate.crt esté en la raíz de tu proyecto y no en .gitignore si es para producción.
# Si tu .env lo referencia y el .env NO se copia al contenedor (buenas práctica para secretos),
# entonces el path en el .env (pasado como variable de entorno a DO) debe ser relativo al WORKDIR aquí.
COPY ca-certificate.crt .

# Exponer el puerto en el que la aplicación Go escucha (ej. 8080)
EXPOSE 8080

# Comando para ejecutar la aplicación
# GIN_MODE=release se debe pasar como variable de entorno en DigitalOcean
ENTRYPOINT ["./yamerito-server"]
