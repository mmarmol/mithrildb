package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidCounterValue = errors.New("document is not a valid counter")
	ErrInvalidCounterType  = errors.New("unsupported value type for counter")
	ErrInvalidDocumentKey  = errors.New("invalid document key")
	ErrInvalidExpiration   = errors.New("invalid expiration value")
)

const (
	DocTypeJSON    = "json"
	DocTypeCounter = "counter"
	DocTypeList    = "list"
	DocTypeSet     = "set"
)

// Metadata holds system-level data associated with a document.
type Metadata struct {
	Rev        string    `json:"rev"`        // Revision ID for conflict resolution
	Expiration int64     `json:"expiration"` // TTL as Unix timestamp (0 = never)
	Type       string    `json:"type"`       // Document type (json, counter, list, set)
	UpdatedAt  time.Time `json:"updated_at"` // When document was last updated
}

// Document is the main object stored in the database.
type Document struct {
	Key   string      `json:"key"`   // Logical key
	Value interface{} `json:"value"` // Content: string, int, []string, map[string]any, etc.
	Meta  Metadata    `json:"meta"`  // Metadata for versioning, expiration, etc.
}

// ValidateValue checks whether a value is valid for the given document type.
func ValidateValue(value interface{}, typeHint string) error {
	switch typeHint {
	case DocTypeJSON:
		_, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
	case DocTypeCounter:
		if _, err := ParseCounterValue(value); err != nil {
			return ErrInvalidCounterValue
		}
	case DocTypeList:
		if _, ok := value.([]interface{}); !ok {
			return errors.New("list value must be a JSON array")
		}
	case DocTypeSet:
		if _, ok := value.([]interface{}); !ok {
			return errors.New("set value must be a JSON array")
		}
	default:
		return fmt.Errorf("unsupported document type: %s", typeHint)
	}
	return nil
}

// ParseCounterValue attempts to convert a document value into an int64 counter.
func ParseCounterValue(val interface{}) (int64, error) {
	switch v := val.(type) {
	case float64:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case string:
		var result int64
		_, err := fmt.Sscanf(v, "%d", &result)
		if err != nil {
			return 0, ErrInvalidCounterValue
		}
		return result, nil
	default:
		return 0, ErrInvalidCounterType
	}
}

var docKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9._:-]{1,250}$`)

// ValidateDocumentKey ensures the key matches naming rules
func ValidateDocumentKey(key string) error {
	if !docKeyRegex.MatchString(key) {
		return fmt.Errorf("%w: invalid characters or length", ErrInvalidDocumentKey)
	}
	if strings.HasPrefix(key, ".") || strings.HasSuffix(key, ".") {
		return fmt.Errorf("%w: key cannot start or end with '.'", ErrInvalidDocumentKey)
	}
	if strings.HasPrefix(key, ":") || strings.HasSuffix(key, ":") {
		return fmt.Errorf("%w: key cannot start or end with ':'", ErrInvalidDocumentKey)
	}
	return nil
}

// IsExpired returns true if the document is expired according to its TTL metadata.
func IsExpired(meta Metadata) bool {
	if meta.Expiration <= 0 {
		return false // 0 o negative "no expiration"
	}
	return time.Now().Unix() > meta.Expiration
}

// ValidateExpiration checks whether an expiration timestamp is valid.
func ValidateExpiration(exp int64) error {
	const maxFutureOffset = 60 * 60 * 24 * 365 * 100 // 100 años
	now := time.Now().Unix()

	switch {
	case exp < 0:
		return ErrInvalidExpiration
	case exp == 0:
		return nil // no expiration
	case exp < now:
		return ErrInvalidExpiration // already expired
	case exp > now+maxFutureOffset:
		return ErrInvalidExpiration // too far in future
	default:
		return nil
	}
}
