package model

import "time"

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
