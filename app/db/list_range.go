package db

import (
	"encoding/json"
	"fmt"
	"mithrildb/model"
)

// GetListRange returns a slice of a list document between the given start and end indices (inclusive).
func (db *DB) GetListRange(opts ListRangeOptions) ([]interface{}, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	if err := model.ValidateDocumentKey(opts.Key); err != nil {
		return nil, err
	}

	if opts.ReadOptions == nil {
		opts.ReadOptions = db.DefaultReadOptions
	}

	val, err := db.TransactionDB.GetCF(opts.ReadOptions, handle, []byte(opts.Key))
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

	start := opts.Start
	end := opts.End

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
