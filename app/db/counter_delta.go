package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// IncrementCounter atomically increments a counter document and returns both the old and new values.
func (db *DB) IncrementCounter(opts CounterIncrementOptions) (oldVal, newVal int64, err error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return 0, 0, ErrInvalidColumnFamily
	}

	if err := model.ValidateDocumentKey(opts.Key); err != nil {
		return 0, 0, err
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

	newVal = oldVal + opts.Delta
	doc.Value = newVal
	doc.Meta.Rev = uuid.NewString()
	doc.Meta.UpdatedAt = time.Now()

	if opts.Expiration != nil {
		if err := model.ValidateExpiration(*opts.Expiration); err != nil {
			txn.Rollback()
			return 0, 0, err
		}
		doc.Meta.Expiration = *opts.Expiration
	}

	data, err := json.Marshal(doc)
	if err != nil {
		txn.Rollback()
		return 0, 0, fmt.Errorf("failed to serialize new counter value: %w", err)
	}

	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return 0, 0, fmt.Errorf("failed to update counter: %w", err)
	}

	if opts.Expiration != nil {
		if err := db.ReplaceTTLInTxn(txn, opts.ColumnFamily, opts.Key, *opts.Expiration); err != nil {
			txn.Rollback()
			return 0, 0, fmt.Errorf("failed to update TTL index: %w", err)
		}
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return oldVal, newVal, nil
}
