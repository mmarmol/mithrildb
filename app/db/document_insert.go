package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// Insert stores a Document only if the key doesn't exist (atomic using transaction).
func (db *DB) Insert(opts PutOptions) (*model.Document, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	err := model.ValidateDocumentKey(opts.Key)
	if err != nil {
		return nil, err
	}

	if opts.Value == nil {
		return nil, ErrNilValue
	}

	if err := model.ValidateValue(opts.Value, opts.Type); err != nil {
		return nil, fmt.Errorf("invalid value for type %s: %w", opts.Type, err)
	}

	// Create default transaction options with snapshot
	txnOpts := grocksdb.NewDefaultTransactionOptions()
	txnOpts.SetSetSnapshot(true)
	defer txnOpts.Destroy()

	// Start transaction
	txn := db.TransactionDB.TransactionBegin(opts.WriteOptions, txnOpts, nil)
	defer txn.Destroy()

	// Set snapshot and read options
	readOpts := grocksdb.NewDefaultReadOptions()
	readOpts.SetSnapshot(txn.GetSnapshot())
	readOpts.SetFillCache(false)
	defer readOpts.Destroy()

	// Read inside the transaction
	val, err := txn.GetWithCF(readOpts, handle, []byte(opts.Key))
	if err != nil {
		txn.Rollback()
		return nil, err
	}
	defer val.Free()

	if val.Exists() {
		raw := val.Data()
		var existing model.Document
		if err := json.Unmarshal(raw, &existing); err == nil {
			if !model.IsExpired(existing.Meta) {
				txn.Rollback()
				return nil, ErrKeyAlreadyExists
			}
		} else {
			txn.Rollback()
			return nil, fmt.Errorf("invalid existing document: %w", err)
		}
	}

	err = model.ValidateExpiration(opts.Expiration)
	if err != nil {
		txn.Rollback()
		return nil, err
	}

	// Build new document
	now := time.Now()
	doc := model.Document{
		Key:   opts.Key,
		Value: opts.Value,
		Meta: model.Metadata{
			Rev:        uuid.NewString(),
			Type:       opts.Type,
			UpdatedAt:  now,
			Expiration: opts.Expiration,
		},
	}

	data, err := json.Marshal(doc)
	if err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to serialize document: %w", err)
	}

	// Put the new document
	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	// Commit the transaction
	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &doc, nil
}
