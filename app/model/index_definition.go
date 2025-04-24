package model

const (
	IndexCFPrefix      = "system.index."
	CFIndexDefinitions = "system.index.definitions"
)

type IndexDefinition struct {
	Name         string   `json:"name"`       // Must be CF-compatible
	Projection   []string `json:"projection"` // Fields used to build index key
	RawCondition string   `json:"condition"`  // Serialized expression (e.g., CEL, expr)

	CreatedAt     int64  `json:"created_at"`      // Creation timestamp
	LastUpdatedAt int64  `json:"last_updated_at"` // Last update applied
	TotalIndexed  uint64 `json:"total_indexed"`   // Total keys indexed
}
