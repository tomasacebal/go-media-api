package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

const (
	passwordAlgorithm  = "pbkdf2_sha256"
	passwordIterations = 120000
	passwordSaltBytes  = 16
	passwordKeyBytes   = 32
)

// HashPassword genera un hash seguro para persistir passwords.
//
// Args:
//   - password: password en texto plano.
//
// Returns:
//   - Hash serializado o error de entropia.
func HashPassword(password string) (string, error) {
	salt := make([]byte, passwordSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generar salt: %w", err)
	}
	key := pbkdf2SHA256([]byte(password), salt, passwordIterations, passwordKeyBytes)
	return strings.Join([]string{
		passwordAlgorithm,
		strconv.Itoa(passwordIterations),
		base64.RawURLEncoding.EncodeToString(salt),
		base64.RawURLEncoding.EncodeToString(key),
	}, "$"), nil
}

// VerifyPassword valida un password contra un hash persistido.
//
// Args:
//   - password: password recibido.
//   - encoded: hash persistido.
//
// Returns:
//   - true si el password coincide.
func VerifyPassword(password string, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != passwordAlgorithm {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}
	salt, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expected, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil || len(expected) == 0 {
		return false
	}
	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func pbkdf2SHA256(password []byte, salt []byte, iterations int, keyLen int) []byte {
	hashLen := sha256.Size
	blocks := (keyLen + hashLen - 1) / hashLen
	output := make([]byte, 0, blocks*hashLen)

	for block := 1; block <= blocks; block++ {
		u := pbkdf2Block(password, salt, iterations, block)
		output = append(output, u...)
	}
	return output[:keyLen]
}

func pbkdf2Block(password []byte, salt []byte, iterations int, block int) []byte {
	blockBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(blockBytes, uint32(block))

	mac := hmac.New(sha256.New, password)
	_, _ = mac.Write(salt)
	_, _ = mac.Write(blockBytes)
	u := mac.Sum(nil)
	result := append([]byte(nil), u...)

	for i := 1; i < iterations; i++ {
		mac = hmac.New(sha256.New, password)
		_, _ = mac.Write(u)
		u = mac.Sum(nil)
		for j := range result {
			result[j] ^= u[j]
		}
	}
	return result
}
