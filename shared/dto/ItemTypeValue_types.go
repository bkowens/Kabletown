package dto

import (
	"encoding/json"
	"errors"
	"strings"
)

// MarshalJSON implements json.Marshaler for ItemTypeValue
func (t ItemTypeValue) MarshalJSON() ([]byte, error) {
	str, ok := ItemTypeValueString[t]
	if !ok {
		// Fallback to numeric value if unknown
		return json.Marshal(int(t))
	}
	return json.Marshal(str)
}

// UnmarshalJSON implements json.Unmarshaler for ItemTypeValue
func (t *ItemTypeValue) UnmarshalJSON(data []byte) error {
	// Try string unmarshaling first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// Case-insensitive matching
		for k, v := range ItemTypeValueByName {
			if strings.EqualFold(k, str) {
				*t = v
				return nil
			}
		}
		// Unknown value - keep as zero/invalid
		return nil
	}

	// Try numeric unmarshaling
	var num int8
	if err := json.Unmarshal(data, &num); err == nil {
		*t = ItemTypeValue(num)
		return nil
	}

	return errors.New("invalid ItemTypeValue")
}
