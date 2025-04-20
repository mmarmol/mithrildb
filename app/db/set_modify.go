package db

import (
	"encoding/json"
	"fmt"
	"mithrildb/events"
	"mithrildb/model"
	"time"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// AddToSet adds an element to a set-type document.
func (db *DB) AddToSet(opts SetOpOptions, element interface{}) (interface{}, error) {
	return db.withSetTransaction(opts, func(set map[interface{}]bool) (map[interface{}]bool, interface{}, error) {
		set[element] = true
		return set, nil, nil
	})
}

// RemoveFromSet removes an element from a set-type document.
func (db *DB) RemoveFromSet(opts SetOpOptions, element interface{}) (interface{}, error) {
	return db.withSetTransaction(opts, func(set map[interface{}]bool) (map[interface{}]bool, interface{}, error) {
		delete(set, element)
		return set, nil, nil
	})
}

// withSetTransaction applies a transactional update to a set document.
func (db *DB) withSetTransaction(
	opts SetOpOptions,
	modifier func(map[interface{}]bool) (map[interface{}]bool, interface{}, error),
) (interface{}, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	if err := model.ValidateDocumentKey(opts.Key); err != nil {
		return nil, err
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

	if model.IsExpired(doc.Meta) {
		txn.Rollback()
		return nil, ErrKeyNotFound
	}

	metaCopy := doc.Meta

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

	newSlice := make([]interface{}, 0, len(newSet))
	for k := range newSet {
		newSlice = append(newSlice, k)
	}

	doc.Value = newSlice
	doc.Meta.Rev = uuid.NewString()
	doc.Meta.UpdatedAt = time.Now()

	if opts.Expiration != nil {
		if err := model.ValidateExpiration(*opts.Expiration); err != nil {
			txn.Rollback()
			return nil, err
		}
		doc.Meta.Expiration = *opts.Expiration
	}

	data, err := json.Marshal(doc)
	if err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to write document: %w", err)
	}

	if err := events.PublishChangeEvent(events.ChangeEventOptions{
		Txn:                txn,
		CFName:             opts.ColumnFamily,
		Key:                opts.Key,
		Document:           &doc,
		Operation:          events.OpMutate,
		PreviousMeta:       &metaCopy,
		ExplicitExpiration: opts.Expiration,
	}); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to enqueue change event: %w", err)
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}
