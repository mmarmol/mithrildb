package db

import (
	"encoding/json"
	"fmt"
	"mithrildb/model"
	"reflect"
)

// CheckSetContains returns true if the given element exists in the set document.
func (db *DB) CheckSetContains(opts SetContainsOptions) (bool, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return false, ErrInvalidColumnFamily
	}

	if opts.ReadOptions == nil {
		opts.ReadOptions = db.DefaultReadOptions
	}

	if err := model.ValidateDocumentKey(opts.Key); err != nil {
		return false, err
	}

	val, err := db.TransactionDB.GetCF(opts.ReadOptions, handle, []byte(opts.Key))
	if err != nil {
		return false, err
	}
	defer val.Free()

	if !val.Exists() {
		return false, ErrKeyNotFound
	}

	var doc model.Document
	if err := json.Unmarshal(val.Data(), &doc); err != nil {
		return false, fmt.Errorf("failed to decode document: %w", err)
	}

	set, ok := doc.Value.([]interface{})
	if !ok {
		return false, ErrInvalidSetType
	}

	for _, item := range set {
		if reflect.DeepEqual(item, opts.Element) {
			return true, nil
		}
	}
	return false, nil
}
