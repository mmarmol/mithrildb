package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// IncrementCounter increments a counter document atomically and returns the old and new values.
func (db *DB) DeltaCounter(cf, key string, delta, expiration int64, opts *grocksdb.WriteOptions) (oldVal, newVal int64, err error) {
	handle, ok := db.Families[cf]
	if !ok {
		return 0, 0, ErrInvalidColumnFamily
	}
	err = model.ValidateDocumentKey(key)
	if err != nil {
		return 0, 0, err
	}

	// Prepare transaction
	txnOpts := grocksdb.NewDefaultTransactionOptions()
	txnOpts.SetSetSnapshot(true)
	defer txnOpts.Destroy()

	txn := db.TransactionDB.TransactionBegin(opts, txnOpts, nil)
	defer txn.Destroy()

	readOpts := grocksdb.NewDefaultReadOptions()
	readOpts.SetSnapshot(txn.GetSnapshot())
	readOpts.SetFillCache(false)
	defer readOpts.Destroy()

	// Read current value
	val, err := txn.GetWithCF(readOpts, handle, []byte(key))
	if err != nil {
		txn.Rollback()
		return 0, 0, err
	}
	defer val.Free()

	var doc model.Document
	if val.Exists() {
		if err := json.Unmarshal(val.Data(), &doc); err != nil {
			txn.Rollback()
			return 0, 0, fmt.Errorf("failed to decode existing counter: %w", err)
		}
		if doc.Meta.Type != model.DocTypeCounter {
			txn.Rollback()
			return 0, 0, model.ErrInvalidCounterType
		}
		if model.IsExpired(doc.Meta) {
			txn.Rollback()
			return 0, 0, ErrKeyNotFound
		}
		oldVal, err = model.ParseCounterValue(doc.Value)
		if err != nil {
			txn.Rollback()
			return 0, 0, err
		}
	} else {
		return 0, 0, ErrKeyNotFound
	}

	err = model.ValidateExpiration(expiration)
	if err != nil {
		txn.Rollback()
		return 0, 0, err
	}
	newVal = oldVal + delta
	now := time.Now()
	doc = model.Document{
		Key:   key,
		Value: newVal,
		Meta: model.Metadata{
			Rev:        uuid.NewString(),
			Type:       model.DocTypeCounter,
			UpdatedAt:  now,
			Expiration: expiration,
		},
	}

	data, err := json.Marshal(doc)
	if err != nil {
		txn.Rollback()
		return 0, 0, fmt.Errorf("failed to serialize new counter value: %w", err)
	}

	if err := txn.PutCF(handle, []byte(key), data); err != nil {
		txn.Rollback()
		return 0, 0, fmt.Errorf("failed to update counter: %w", err)
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return oldVal, newVal, nil
}
