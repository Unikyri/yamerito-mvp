# Etapa 1: Builder - Compilar Frontend y Backend
FROM golang:1.24.1-alpine AS builder

# Instalar dependencias para frontend (Node.js, npm) y backend (git, gcc, musl-dev para CGO si es necesario)
RUN apk add --no-cache nodejs npm git gcc musl-dev

# Establecer directorio de trabajo para toda la aplicación
WORKDIR /app

# --- Construir Frontend ---
# Copiar archivos de configuración del frontend primero para cachear dependencias de npm
COPY frontend/package.json frontend/package-lock.json* ./frontend/
WORKDIR /app/frontend
RUN npm install

# Copiar el resto de los archivos del frontend y construir
COPY frontend/ ./ 
RUN npm run build
# Los archivos compilados del frontend estarán en /app/frontend/dist

# --- Construir Backend ---
# Volver al directorio raíz de la aplicación
WORKDIR /app

# Copiar archivos go.mod y go.sum para cachear dependencias de Go
COPY go.mod go.sum ./
# Descargar dependencias de Go (opcional, go build también las descarga)
# RUN go mod download 

# Copiar el resto del código fuente de la aplicación Go
COPY . .

# Compilar el backend Go (nuestro servidor API)
# El output será /app/yamerito-server
RUN go build -ldflags="-s -w" -o yamerito-server ./cmd/server/main.go


# Etapa 2: Final - Crear la imagen de producción ligera
FROM alpine:latest AS final

# Argumento para el puerto (puede ser sobrescrito en tiempo de ejecución de DigitalOcean)
ARG APP_PORT=8080
ENV PORT=${APP_PORT}

WORKDIR /app

# Copiar el binario del backend compilado desde la etapa builder
COPY --from=builder /app/yamerito-server .

# Copiar los assets del frontend compilado desde la etapa builder
# El servidor Go espera encontrarlos en ./frontend/dist relativo a su ubicación
COPY --from=builder /app/frontend/dist ./frontend/dist/

# Copiar el certificado CA si es necesario para la conexión a la BD en producción
# Asegúrate que ca-certificate.crt esté en la raíz de tu proyecto y NO en .gitignore
# Si no usas SSL o la CA es pública, puedes omitir esto.
COPY ca-certificate.crt ./

# Exponer el puerto en el que la aplicación se ejecutará
EXPOSE ${APP_PORT}

# Comando para ejecutar la aplicación
# El backend (yamerito-server) ahora sirve la API y el frontend
ENTRYPOINT ["/app/yamerito-server"]
