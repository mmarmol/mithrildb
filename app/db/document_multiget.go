package db

import (
	"encoding/json"
	"fmt"

	"mithrildb/model"

	"github.com/linxGnu/grocksdb"
)

// MultiGet retrieves multiple documents by keys from a specific column family.
func (db *DB) MultiGet(cf string, keys []string, opts *grocksdb.ReadOptions) (map[string]*model.Document, error) {
	handle, ok := db.Families[cf]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	if len(keys) == 0 {
		return map[string]*model.Document{}, nil
	}

	byteKeys := make([][]byte, len(keys))
	for i, k := range keys {
		err := model.ValidateDocumentKey(k)
		if err != nil {
			return nil, err
		}
		byteKeys[i] = []byte(k)
	}

	values, err := db.TransactionDB.MultiGetWithCF(opts, handle, byteKeys...)
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

	result := make(map[string]*model.Document, len(keys))
	for i, val := range values {
		if val == nil || val.Size() == 0 {
			result[keys[i]] = nil
		} else {
			var doc model.Document
			if err := json.Unmarshal(val.Data(), &doc); err != nil {
				return nil, fmt.Errorf("failed to decode document for key '%s': %w", keys[i], err)
			}
			if model.IsExpired(doc.Meta) {
				result[keys[i]] = nil
			} else {
				result[keys[i]] = &doc
			}
		}
	}

	return result, nil
}
