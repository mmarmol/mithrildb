package db

import (
	"encoding/json"
	"fmt"

	"mithrildb/model"
)

// GetDocument retrieves a full document by key from a specific column family.
func (db *DB) GetDocument(opts DocumentReadOptions) (*model.Document, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	if err := model.ValidateDocumentKey(opts.Key); err != nil {
		return nil, err
	}

	value, err := db.TransactionDB.GetCF(opts.ReadOptions, handle, []byte(opts.Key))
	if err != nil {
		return nil, err
	}
	defer value.Free()

	if !value.Exists() || value.Size() == 0 {
		return nil, ErrKeyNotFound
	}

	var doc model.Document
	if err := json.Unmarshal(value.Data(), &doc); err != nil {
		return nil, fmt.Errorf("failed to decode stored document: %w", err)
	}

	if model.IsExpired(doc.Meta) {
		return nil, ErrKeyNotFound
	}

	return &doc, nil
}
