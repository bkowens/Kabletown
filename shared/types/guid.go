package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// GUIDRegex matches standard UUID format (with or without hyphens)
var GUIDRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// GUIDRegexNoHyphens matches UUID format without hyphens
var GUIDRegexNoHyphens = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)

// GUIDRegex32 matches UUID format without hyphens
var GUIDRegex32 = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)

// IsValidGUID checks if a string is a valid GUID/UUID
func IsValidGUID(guid string) bool {
	if guid == "" {
		return false
	}
	
	// Try with hyphens
	if GUIDRegex.MatchString(guid) {
		return true
	}
	
	// Try without hyphens
	if GUIDRegex32.MatchString(guid) {
		return true
	}
	
	return false
}

// ParseGUID parses a GUID string and returns a uuid.UUID
// Accepts both hyphenated and non-hyphenated formats
func ParseGUID(guid string) (uuid.UUID, error) {
	if guid == "" {
		return uuid.Nil, fmt.Errorf("empty GUID")
	}
	
	// Try standard parse first (with hyphens)
	id, err := uuid.Parse(guid)
	if err == nil {
		return id, nil
	}
	
	// Try without hyphens
	if GUIDRegex32.MatchString(guid) {
		formatted := InsertHyphens(guid)
		return uuid.Parse(formatted)
	}
	
	return uuid.Nil, fmt.Errorf("invalid GUID format: %s", guid)
}

// MustParseGUID parses a GUID and panics if invalid
// Useful for compile-time constants
func MustParseGUID(guid string) uuid.UUID {
	id, err := ParseGUID(guid)
	if err != nil {
		panic(fmt.Sprintf("invalid GUID: %s", err))
	}
	return id
}

// NewGUID generates a new random UUID
func NewGUID() string {
	return uuid.New().String()
}

// InsertHyphens adds hyphens to a 32-character hex string to make it a proper UUID
func InsertHyphens(guid string) string {
	if len(guid) != 32 {
		return guid
	}
	
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		guid[0:8],
		guid[8:12],
		guid[12:16],
		guid[16:20],
		guid[20:32],
	)
}

// RemoveHyphens removes hyphens from a UUID string
func RemoveHyphens(guid string) string {
	return strings.ReplaceAll(guid, "-", "")
}

// NormalizeGUID returns a GUID in standard hyphenated format
// Accepts both formats and always returns hyphenated
func NormalizeGUID(guid string) (string, error) {
	id, err := ParseGUID(guid)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// GUIDsMatch compares two GUID strings for equality
// Handles both hyphenated and non-hyphenated formats
func GUIDsMatch(guid1, guid2 string) (bool, error) {
	id1, err1 := ParseGUID(guid1)
	if err1 != nil {
		return false, err1
	}
	
	id2, err2 := ParseGUID(guid2)
	if err2 != nil {
		return false, err2
	}
	
	return id1 == id2, nil
}

// MustGUIDsMatch compares two GUID strings and panics if either is invalid
func MustGUIDsMatch(guid1, guid2 string) bool {
	match, _ := GUIDsMatch(guid1, guid2)
	return match
}

// GUIDSlice is a custom type for JSON marshaling of []uuid.UUID
type GUIDSlice []string

// ContainsGUID checks if a slice of GUID strings contains a specific GUID
func ContainsGUID(guids []string, target string) bool {
	for _, g := range guids {
		match, err := GUIDsMatch(g, target)
		if err == nil && match {
			return true
		}
	}
	return false
}

// UniqueGUIDs removes duplicate GUIDs from a slice
func UniqueGUIDs(guids []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, g := range guids {
		normalized, err := NormalizeGUID(g)
		if err != nil {
			continue
		}
		
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}
	
	return result
}

// ParseGUIDs parses a comma-separated list of GUIDs
func ParseGUIDs(guidList string) ([]uuid.UUID, error) {
	if guidList == "" {
		return []uuid.UUID{}, nil
	}
	
	parts := SplitString(guidList, ",")
	var result []uuid.UUID
	
	for _, part := range parts {
		part = TrimString(part)
		if part == "" {
			continue
		}
		
		id, err := ParseGUID(part)
		if err != nil {
			return nil, fmt.Errorf("invalid GUID '%s': %w", part, err)
		}
		
		result = append(result, id)
	}
	
	return result, nil
}

// GUIDToBytes converts a GUID to a 16-byte array
func GUIDToBytes(guid string) ([]byte, error) {
	id, err := ParseGUID(guid)
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

// BytesToGUID converts a 16-byte array to a GUID string
func BytesToGUID(data []byte) (string, error) {
	if len(data) != 16 {
		return "", fmt.Errorf("invalid byte length: expected 16, got %d", len(data))
	}
	
	var id uuid.UUID
	copy(id[:], data)
	return id.String(), nil
}

// EmptyGUID returns the empty UUID (00000000-0000-0000-0000-000000000000)
func EmptyGUID() string {
	return uuid.Nil.String()
}

// Helper functions to avoid dependency on strings package
func SplitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func TrimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
