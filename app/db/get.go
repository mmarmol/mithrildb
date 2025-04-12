package db

import (
	"encoding/json"
	"fmt"

	"mithrildb/model"

	"github.com/linxGnu/grocksdb"
)

// Get retrieves a full Document by key from a specific column family.
func (db *DB) Get(cf, key string, opts *grocksdb.ReadOptions) (*model.Document, error) {
	handle, ok := db.Families[cf]
	if !ok {
		return nil, fmt.Errorf("column family '%s' does not exist", cf)
	}

	value, err := db.TransactionDB.GetCF(opts, handle, []byte(key))
	if err != nil {
		return nil, err
	}
	defer value.Free()

	if value.Size() == 0 {
		return nil, nil // Key not found
	}

	var doc model.Document
	if err := json.Unmarshal(value.Data(), &doc); err != nil {
		return nil, fmt.Errorf("failed to decode stored document: %w", err)
	}

	return &doc, nil
}
