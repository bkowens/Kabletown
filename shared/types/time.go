package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JellyfinTime represents time values in Jellyfin's 100-ns tick format
type JellyfinTime struct {
	time.Time
}

// NewJellyfinTime creates a new JellyfinTime from time.Time
func NewJellyfinTime(t time.Time) JellyfinTime {
	return JellyfinTime{Time: t}
}

// ToTicks converts JellyfinTime to 100-ns ticks (int64)
func (jt JellyfinTime) ToTicks() int64 {
	// Unix time in nanoseconds, converted to 100-ns units
	return jt.UnixNano() / 100
}

// NewJellyfinTimeFromTicks creates a JellyfinTime from 100-ns ticks
func NewJellyfinTimeFromTicks(ticks int64) JellyfinTime {
	// Convert 100-ns ticks to nanoseconds, then to time.Time
	nano := ticks * 100
	return JellyfinTime{Time: time.Unix(0, nano)}
}

// MarshalJSON implements json.Marshaler for JellyfinTime
func (jt JellyfinTime) MarshalJSON() ([]byte, error) {
	if jt.IsZero() {
		return []byte("null"), nil
	}
	// Serialize as ticks (int64)
	return json.Marshal(jt.ToTicks())
}

// UnmarshalJSON implements json.Unmarshaler for JellyfinTime
func (jt *JellyfinTime) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as ticks (int64)
	var ticks int64
	if err := json.Unmarshal(data, &ticks); err == nil {
		*jt = NewJellyfinTimeFromTicks(ticks)
		return nil
	}
	// Fall back to string parsing (ISO 8601)
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*jt = JellyfinTime{Time: parsed}
	return nil
}

// MarshalText implements encoding.TextMarshaler for JellyfinTime
func (jt JellyfinTime) MarshalText() ([]byte, error) {
	if jt.IsZero() {
		return nil, nil
	}
	return []byte(fmt.Sprintf("%d", jt.ToTicks())), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JellyfinTime
func (jt *JellyfinTime) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*jt = JellyfinTime{Time: time.Time{}}
		return nil
	}
	var ticks int64
	_, err := fmt.Sscanf(string(data), "%d", &ticks)
	if err != nil {
		return err
	}
	*jt = NewJellyfinTimeFromTicks(ticks)
	return nil
}

// Value implements driver.Valuer for database storage
func (jt JellyfinTime) Value() (driver.Value, error) {
	if jt.IsZero() {
		return nil, nil
	}
	return jt.ToTicks(), nil
}

// Scan implements sql.Scanner for database retrieval
func (jt *JellyfinTime) Scan(value interface{}) error {
	if value == nil {
		*jt = JellyfinTime{Time: time.Time{}}
		return nil
	}
	switch v := value.(type) {
	case int64:
		*jt = NewJellyfinTimeFromTicks(v)
		return nil
	case []byte:
		var ticks int64
		n_, err := fmt.Sscanf(string(v), "%d", &ticks)
		if err != nil && n_ == 0 {
			return err
		}
		*jt = NewJellyfinTimeFromTicks(ticks)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into JellyfinTime", value)
	}
}

// CompareComparison compares two JellyfinTimes for sorting
func (jt JellyfinTime) Compare(other JellyfinTime) int {
	return jt.Time.Compare(other.Time)
}

// Before returns whether jt occurs before other
func (jt JellyfinTime) Before(other JellyfinTime) bool {
	return jt.Time.Before(other.Time)
}

// After returns whether jt occurs after other
func (jt JellyfinTime) After(other JellyfinTime) bool {
	return jt.Time.After(other.Time)
}

// Zero returns whether the JellyfinTime represents a zero time
func (jt JellyfinTime) IsZero() bool {
	return jt.Time.IsZero()
}

// AddDuration adds a time.Duration to this JellyfinTime
func (jt JellyfinTime) AddDuration(d time.Duration) JellyfinTime {
	return JellyfinTime{Time: jt.Time.Add(d)}
}

// SubTicks subtracts another JellyfinTime, returning the difference in ticks
func (jt JellyfinTime) SubTicks(other JellyfinTime) int64 {
	return jt.ToTicks() - other.ToTicks()
}

// ParseJellyfinTime parses a Jellyfin-formatted time string
func ParseJellyfinTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized time format: %s", timeStr)
}