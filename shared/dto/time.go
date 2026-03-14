// Package dto contains shared data transfer objects used across Kabletown services.
package dto

import (
	"encoding/json"
	"time"
)

// JellyfinTime represents a time that can be serialized as DateTime, DateTimeNullable, or Empty
type JellyfinTime struct {
	time.Time
	IsNull bool
}

// NewJellyfinTime creates a new JellyfinTime from a time.Time
func NewJellyfinTime(t time.Time) JellyfinTime {
	return JellyfinTime{Time: t, IsNull: false}
}

// NewJellyfinTimePointer creates a pointer to a JellyfinTime from a time.Time
func NewJellyfinTimePointer(t time.Time) *JellyfinTime {
	return &JellyfinTime{Time: t, IsNull: false}
}

// NewNullJellyfinTime creates a null JellyfinTime
func NewNullJellyfinTime() JellyfinTime {
	return JellyfinTime{Time: time.Time{}, IsNull: true}
}

// NewNullJellyfinTimePointer creates a null JellyfinTime pointer
func NewNullJellyfinTimePointer() *JellyfinTime {
	return &JellyfinTime{Time: time.Time{}, IsNull: true}
}

// MarshalJSON implements json.Marshaler for JellyfinTime
func (j JellyfinTime) MarshalJSON() ([]byte, error) {
	if j.IsNull {
		return []byte("null"), nil
	}
		s := j.Time.Format("2006-01-02T15:04:05.0000000Z")
	return json.Marshal(s)
}

// UnmarshalJSON implements json.Unmarshaler for JellyfinTime
func (j *JellyfinTime) UnmarshalJSON(data []byte) error {
	j.IsNull = false
	
	// Check for null
	if string(data) == "null" {
		j.IsNull = true
		return nil
	}
	
	// Parse string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	
	// Try standard formats
	if t, err := time.Parse("2006-01-02T15:04:05.0000000Z", s); err == nil {
		j.Time = t
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		j.Time = t
		return nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		j.Time = t
		return nil
	}
	
	// Default: RFC3339 with timezone
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		j.IsNull = true
		return nil // Don't fail on invalid date, just mark as null
	}
	
	j.Time = t
	return nil
}

// MarshalJSONPtr is a helper to marshaling *JellyfinTime
func (j *JellyfinTime) MarshalJSONPtr() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*j)
}

// IsZero returns true if the time is zero/empty
func (j JellyfinTime) IsZero() bool {
	return j.IsNull || j.Time.IsZero()
}

// ToTime returns the underlying time.Time, or Zero time if null
func (j JellyfinTime) ToTime() time.Time {
	if j.IsNull {
		return time.Time{}
	}
	return j.Time
}

// FromTime creates a JellyfinTime from time.Time
func FromTime(t time.Time) JellyfinTime {
	return JellyfinTime{Time: t, IsNull: t.IsZero()}
}

// FromNullableTime creates a JellyfinTime from a nullable time pointer
func FromNullableTime(t *time.Time) JellyfinTime {
	if t == nil || t.IsZero() {
		return JellyfinTime{IsNull: true}
	}
	return JellyfinTime{Time: *t, IsNull: false}
}
