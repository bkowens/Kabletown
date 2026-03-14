package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// GenerateToken creates a 256-bit (64 hex char) token for API authentication
func GenerateToken() string {
	token := make([]byte, 32) // 32 bytes = 256 bits
	rand.Read(token)
	return hex.EncodeToString(token)
}

// HashToken creates SHA256 hash of token for secure database storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// TokenExpiry checks if a token has expired
func TokenExpiry(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false // Indefinite lifetime
	}
	return time.Now().After(*expiresAt)
}

// ValidateTokenFormat checks if token is valid 64-char hex
func ValidateTokenFormat(token string) bool {
	if len(token) != 64 {
		return false
	}
	_, err := hex.DecodeString(token)
	return err == nil
}
