package types

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TokenLength is the number of bytes in a token (256-bit = 32 bytes = 64 hex chars)
const TokenLength = 32

// GenerateToken creates a new 256-bit hex token
func GenerateToken() string {
	bytes := make([]byte, TokenLength)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// HashToken creates SHA256 hash of a token for DB storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ValidateTokenFormat checks if a token has valid format
func ValidateTokenFormat(token string) bool {
	if len(token) != 64 {
		return false
	}
	
	_, err := hex.DecodeString(token)
	return err == nil
}
