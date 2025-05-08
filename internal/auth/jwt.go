package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/Unikyri/yamerito-mvp/internal/models" // Para usar models.Role
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey []byte

// InitJWT inicializa la clave secreta JWT desde las variables de entorno.
// Debe llamarse una vez al inicio de la aplicación.
func InitJWT() error {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return errors.New("JWT_SECRET_KEY no está configurada en las variables de entorno")
	}
	jwtSecretKey = []byte(secret)
	return nil
}

// Claims define la estructura de los claims para el token JWT.
type Claims struct {
	UserID   uint        `json:"user_id"`
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT genera un nuevo token JWT para un usuario.
func GenerateJWT(userID uint, username string, role models.Role) (string, error) {
	if len(jwtSecretKey) == 0 {
		// Esto es un fallback, idealmente InitJWT() ya fue llamado y manejó el error.
		if err := InitJWT(); err != nil {
			return "", err
		}
	}

	// Definir el tiempo de expiración del token (ej. 24 horas)
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "yamerito-mvp", // Puedes cambiar el emisor
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT valida un token JWT y devuelve los claims si es válido.
func ValidateJWT(tokenString string) (*Claims, error) {
	if len(jwtSecretKey) == 0 {
		// Asegurarse de que InitJWT haya sido llamado
		if err := InitJWT(); err != nil {
			return nil, err
		}
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verificar el método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inesperado")
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expirado")
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, errors.New("token aún no es válido")
		} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, errors.New("firma de token inválida")
		} else if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, errors.New("token malformado")
		}
		// Otro error de parseo no específico
		log.Printf("Error al parsear token: %v", err) // Registrar el error real
		return nil, errors.New("token inválido o error de parseo")
	}

	if !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}
