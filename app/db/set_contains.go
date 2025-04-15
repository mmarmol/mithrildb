package db

import (
	"encoding/json"
	"fmt"
	"mithrildb/model"
	"reflect"

	"github.com/linxGnu/grocksdb"
)

func (db *DB) SetContains(cf string, key string, element interface{}, opts *grocksdb.ReadOptions) (bool, error) {
	handle, ok := db.Families[cf]
	if !ok {
		return false, ErrInvalidColumnFamily
	}

	if opts == nil {
		opts = db.DefaultReadOptions
	}

	err := model.ValidateDocumentKey(key)
	if err != nil {
		return false, err
	}

	val, err := db.TransactionDB.GetCF(opts, handle, []byte(key))
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
		if reflect.DeepEqual(item, element) {
			return true, nil
		}
	}
	return false, nil
}
