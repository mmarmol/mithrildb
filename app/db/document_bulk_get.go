package db

import (
	"encoding/json"
	"fmt"

	"mithrildb/model"
)

// BulkGetDocuments retrieves multiple documents by key from the given column family.
func (db *DB) BulkGetDocuments(opts BulkReadOptions) (map[string]*model.Document, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	if len(opts.Keys) == 0 {
		return map[string]*model.Document{}, nil
	}

	byteKeys := make([][]byte, len(opts.Keys))
	for i, k := range opts.Keys {
		if err := model.ValidateDocumentKey(k); err != nil {
			return nil, err
		}
		byteKeys[i] = []byte(k)
	}

	values, err := db.TransactionDB.MultiGetWithCF(opts.ReadOptions, handle, byteKeys...)
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, val := range values {
			if val != nil {
				val.Free()
			}
		}
	}()

	result := make(map[string]*model.Document, len(opts.Keys))
	for i, val := range values {
		if val == nil || val.Size() == 0 {
			result[opts.Keys[i]] = nil
		} else {
			var doc model.Document
			if err := json.Unmarshal(val.Data(), &doc); err != nil {
				return nil, fmt.Errorf("failed to decode document for key '%s': %w", opts.Keys[i], err)
			}
			if model.IsExpired(doc.Meta) {
				result[opts.Keys[i]] = nil
			} else {
				result[opts.Keys[i]] = &doc
			}
		}
	}

	return result, nil
}
