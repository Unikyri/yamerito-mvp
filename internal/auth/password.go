package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2idParams define los parámetros para el hashing Argon2id.
// Estos son valores recomendados por OWASP. Ajusta según tus necesidades y pruebas de rendimiento.
type Argon2idParams struct {
	Memory      uint32 // Memoria en KiB
	Iterations  uint32 // Número de iteraciones
	Parallelism uint8  // Grado de paralelismo (número de hilos)
	SaltLength  uint32 // Longitud del salt en bytes
	KeyLength   uint32 // Longitud de la clave derivada (hash) en bytes
}

// DefaultParams son los parámetros por defecto y recomendados para Argon2id.
var DefaultParams = &Argon2idParams{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// HashPassword genera un hash Argon2id para la contraseña dada.
// El formato del hash devuelto es: $argon2id$v=19$m=<memory>,t=<iterations>,p=<parallelism>$<salt_base64>$<hash_base64>
func HashPassword(password string, p *Argon2idParams) (string, error) {
	if p == nil {
		p = DefaultParams
	}

	// Generar un salt aleatorio
	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("error al generar salt: %w", err)
	}

	// Derivar la clave (hash)
	hash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	// Codificar salt y hash en Base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Formatear la cadena de hash
	// Sigue el formato estándar para que sea compatible con otras implementaciones
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Iterations, p.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// CheckPasswordHash verifica una contraseña en texto plano contra un hash Argon2id almacenado.
func CheckPasswordHash(password, encodedHash string) (match bool, err error) {
	// Parsear el hash almacenado
	p, salt, hash, err := DecodeHash(encodedHash)
	if err != nil {
		return false, fmt.Errorf("error al decodificar hash: %w", err)
	}

	// Derivar el hash de la contraseña proporcionada usando los mismos parámetros
	otherHash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	// Comparar los hashes en tiempo constante para prevenir ataques de temporización
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

// DecodeHash parsea una cadena de hash Argon2id y extrae los parámetros, salt y hash.
func DecodeHash(encodedHash string) (p *Argon2idParams, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("formato de hash inválido: número incorrecto de partes")
	}

	if vals[1] != "argon2id" {
		return nil, nil, nil, errors.New("formato de hash inválido: no es argon2id")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil || version != argon2.Version {
		return nil, nil, nil, errors.New("formato de hash inválido: versión incompatible")
	}

	p = &Argon2idParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("formato de hash inválido: error al parsear parámetros: %w", err)
	}

	salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("formato de hash inválido: error al decodificar salt: %w", err)
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("formato de hash inválido: error al decodificar hash: %w", err)
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
