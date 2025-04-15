package db

import (
	"encoding/json"
	"fmt"
	"mithrildb/model"

	"github.com/linxGnu/grocksdb"
)

// ListRange devuelve una porción de la lista entre índices start y end (inclusive)
func (db *DB) ListRange(cf string, key string, start, end int, opts *grocksdb.ReadOptions) ([]interface{}, error) {
	handle, ok := db.Families[cf]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	if opts == nil {
		opts = db.DefaultReadOptions
	}

	val, err := db.TransactionDB.GetCF(opts, handle, []byte(key))
	if err != nil {
		return nil, err
	}
	defer val.Free()

	if !val.Exists() {
		return nil, ErrKeyNotFound
	}

	var doc model.Document
	if err := json.Unmarshal(val.Data(), &doc); err != nil {
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	list, ok := doc.Value.([]interface{})
	if !ok {
		return nil, ErrInvalidListType
	}

	if start < 0 {
		start = 0
	}
	if end < 0 || end >= len(list) {
		end = len(list) - 1
	}
	if start > end {
		return []interface{}{}, nil
	}

	return list[start : end+1], nil
}
