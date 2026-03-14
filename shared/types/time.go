package types

import (
	"encoding/json"
	"time"
)

// JellyfinTime ensures timestamps are serialized with exactly 7 decimal places (W3)
// Format: 2006-01-02T15:04:05.0000000Z
type JellyfinTime struct {
	time.Time
}

// Now returns the current UTC time as a JellyfinTime
func Now() JellyfinTime {
	return JellyfinTime{Time: time.Now().UTC()}
}

// FromTime creates a JellyfinTime from a time.Time
func FromTime(t time.Time) JellyfinTime {
	return JellyfinTime{Time: t.UTC()}
}

// MarshalJSON encodes the time with exactly 7 decimal places in UTC
// Returns null for zero values
func (jt JellyfinTime) MarshalJSON() ([]byte, error) {
	if jt.IsZero() {
		return []byte("null"), nil
	}
	ts := jt.UTC().Format("2006-01-02T15:04:05.0000000Z")
	return json.Marshal(ts)
}

// UnmarshalJSON decodes time with 7 decimal places
func (jt *JellyfinTime) UnmarshalJSON(b []byte) error {
	// Handle null
	if string(b) == "null" {
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02T15:04:05.0000000Z", s)
	if err != nil {
		return err
	}
	jt.Time = t
	return nil
}

// Ptr returns a pointer to the JellyfinTime
func (jt JellyfinTime) Ptr() *JellyfinTime {
	return &jt
}
