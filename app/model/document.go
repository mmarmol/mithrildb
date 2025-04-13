package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidCounterValue = errors.New("document is not a valid counter")
	ErrInvalidCounterType  = errors.New("unsupported value type for counter")
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
	CreatedAt  time.Time `json:"created_at"` // When document was created
	UpdatedAt  time.Time `json:"updated_at"` // When document was last updated
}

// Document is the main object stored in the database.
type Document struct {
	Key   string      `json:"key"`   // Logical key
	Value interface{} `json:"value"` // Content: string, int, []string, map[string]any, etc.
	Meta  Metadata    `json:"meta"`  // Metadata for versioning, expiration, etc.
}

func ValidateValue(value interface{}, typeHint string) error {
	switch typeHint {
	case DocTypeJSON:
		// Debe poder serializar a JSON
		_, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
	case DocTypeCounter:
		// Debe ser entero (int, int64, float64, etc)
		if _, err := ParseCounterValue(value); err != nil {
			return ErrInvalidCounterValue
		}
	case DocTypeList:
		// Debe ser slice
		if _, ok := value.([]interface{}); !ok {
			return errors.New("list value must be a JSON array")
		}
	case DocTypeSet:
		// También array, pero sin repetidos (eso se puede validar más adelante)
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
