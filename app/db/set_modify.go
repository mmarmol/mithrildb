package db

import (
	"encoding/json"
	"fmt"
	"mithrildb/model"
	"time"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

type SetOpOptions = ListOpOptions

func (db *DB) withSetTransaction(opts SetOpOptions, modifier func(map[interface{}]bool) (map[interface{}]bool, interface{}, error)) (interface{}, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	err := model.ValidateDocumentKey(opts.Key)
	if err != nil {
		return false, err
	}

	txnOpts := grocksdb.NewDefaultTransactionOptions()
	txnOpts.SetSetSnapshot(true)
	defer txnOpts.Destroy()

	txn := db.TransactionDB.TransactionBegin(opts.WriteOptions, txnOpts, nil)
	defer txn.Destroy()

	readOpts := grocksdb.NewDefaultReadOptions()
	readOpts.SetSnapshot(txn.GetSnapshot())
	readOpts.SetFillCache(false)
	defer readOpts.Destroy()

	val, err := txn.GetWithCF(readOpts, handle, []byte(opts.Key))
	if err != nil {
		txn.Rollback()
		return nil, err
	}
	defer val.Free()

	if !val.Exists() {
		txn.Rollback()
		return nil, ErrKeyNotFound
	}

	var doc model.Document
	if err := json.Unmarshal(val.Data(), &doc); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	slice, ok := doc.Value.([]interface{})
	if !ok {
		txn.Rollback()
		return nil, ErrInvalidSetType
	}

	set := make(map[interface{}]bool)
	for _, v := range slice {
		set[v] = true
	}

	newSet, result, err := modifier(set)
	if err != nil {
		txn.Rollback()
		return nil, err
	}

	// convert map back to slice
	newSlice := make([]interface{}, 0, len(newSet))
	for k := range newSet {
		newSlice = append(newSlice, k)
	}

	doc.Value = newSlice
	doc.Meta.Rev = uuid.NewString()
	doc.Meta.UpdatedAt = time.Now()
	doc.Meta.Expiration = opts.Expiration

	data, err := json.Marshal(doc)
	if err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to write document: %w", err)
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// SetAdd agrega un elemento al set si no existe
func (db *DB) SetAdd(opts SetOpOptions, element interface{}) (interface{}, error) {
	return db.withSetTransaction(opts, func(set map[interface{}]bool) (map[interface{}]bool, interface{}, error) {
		set[element] = true
		return set, nil, nil
	})
}

// SetRemove elimina un elemento del set
func (db *DB) SetRemove(opts SetOpOptions, element interface{}) (interface{}, error) {
	return db.withSetTransaction(opts, func(set map[interface{}]bool) (map[interface{}]bool, interface{}, error) {
		delete(set, element)
		return set, nil, nil
	})
}
