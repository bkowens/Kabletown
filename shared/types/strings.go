package types

import "strings"

// SplitPipeDelimited converts "a|b|c" to []string{"a", "b", "c"}
// Nullable version: handles *string from database (W7 compliant)
func SplitPipeDelimited(s *string) []string {
    if s == nil || *s == "" || *s == "|" {
        return []string{}  // W6: Empty array, not null
    }
    
    parts := strings.Split(*s, "|")
    result := make([]string, 0, len(parts))
    
    for _, p := range parts {
        if p != "" {
            result = append(result, p)
        }
    }
    
    return result
}

// SplitPipeDelimitedString handles non-pointer strings
func SplitPipeDelimitedString(s string) []string {
    if s == "" || s == "|" {
        return []string{}
    }
    
    parts := strings.Split(s, "|")
    result := make([]string, 0, len(parts))
    
    for _, p := range parts {
        if p != "" {
            result = append(result, p)
        }
    }
    
    return result
}

// JoinPipeDelimited converts []string{"a", "b", "c"} to "a|b|c"
func JoinPipeDelimited(parts []string) string {
    if len(parts) == 0 {
        return ""
    }
    return strings.Join(parts, "|")
}
