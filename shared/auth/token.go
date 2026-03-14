package auth

import "github.com/jellyfinhanced/shared/types"

// TokenLength is the number of bytes in a token (256-bit = 32 bytes = 64 hex chars)
const TokenLength = types.TokenLength

// GenerateToken creates a new 256-bit hex token  
func GenerateToken() string {
	return types.GenerateToken()
}

// HashToken creates SHA256 hash of a token for DB storage
func HashToken(token string) string {
	return types.HashToken(token)
}

// ValidateTokenFormat checks if a token has valid format
func ValidateTokenFormat(token string) bool {
	return types.ValidateTokenFormat(token)
}
