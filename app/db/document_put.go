package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// PutOptions defines configurable parameters for a document insert or update.
type PutOptions struct {
	ColumnFamily string                 // Target column family
	Key          string                 // Document key
	Value        interface{}            // Document value (string, map, etc.)
	Cas          string                 // Optional revision for optimistic locking
	Type         string                 // Document type (json, counter, etc.)
	Expiration   *int64                 // Optional TTL (Unix timestamp)
	WriteOptions *grocksdb.WriteOptions // RocksDB write options
}

// Put stores a Document using the provided options, handling revision control.
func (db *DB) PutWithOptions(opts PutOptions) (*model.Document, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	if err := model.ValidateDocumentKey(opts.Key); err != nil {
		return nil, err
	}
	if opts.Value == nil {
		return nil, ErrNilValue
	}
	if err := model.ValidateValue(opts.Value, opts.Type); err != nil {
		return nil, fmt.Errorf("invalid value for type %s: %w", opts.Type, err)
	}
	if opts.Expiration != nil {
		if err := model.ValidateExpiration(*opts.Expiration); err != nil {
			return nil, err
		}
	}

	// Always use transaction now
	txnOpts := grocksdb.NewDefaultTransactionOptions()
	txnOpts.SetSetSnapshot(true)
	txnOpts.SetLockTimeout(1000)
	defer txnOpts.Destroy()

	txn := db.TransactionDB.TransactionBegin(opts.WriteOptions, txnOpts, nil)
	defer txn.Destroy()

	// Check CAS if provided
	if opts.Cas != "" {
		readOpts := grocksdb.NewDefaultReadOptions()
		readOpts.SetFillCache(false)
		defer readOpts.Destroy()

		val, err := txn.GetWithCF(readOpts, handle, []byte(opts.Key))
		if err != nil {
			txn.Rollback()
			return nil, err
		}
		defer val.Free()

		if val.Size() > 0 {
			var existing model.Document
			if err := json.Unmarshal(val.Data(), &existing); err != nil {
				txn.Rollback()
				return nil, fmt.Errorf("failed to unmarshal existing document: %w", err)
			}
			if existing.Meta.Rev != opts.Cas {
				txn.Rollback()
				return nil, ErrRevisionMismatch
			}
		}
	}

	// Prepare new document
	now := time.Now()
	expiration := int64(0)
	if opts.Expiration != nil {
		expiration = *opts.Expiration
	}

	doc := model.Document{
		Key:   opts.Key,
		Value: opts.Value,
		Meta: model.Metadata{
			Rev:        uuid.NewString(),
			Type:       opts.Type,
			UpdatedAt:  now,
			Expiration: expiration,
		},
	}

	data, err := json.Marshal(doc)
	if err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("error marshaling document: %w", err)
	}

	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to put document: %w", err)
	}

	if opts.Expiration != nil {
		if err := db.ReplaceTTLInTxn(txn, opts.ColumnFamily, opts.Key, *opts.Expiration); err != nil {
			txn.Rollback()
			return nil, fmt.Errorf("failed to update TTL index: %w", err)
		}
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &doc, nil
}
