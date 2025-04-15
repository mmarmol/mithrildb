package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// Replace stores a Document only if the key already exists (atomic using transaction).
func (db *DB) Replace(opts PutOptions) (*model.Document, error) {
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

	raw := val.Data()
	var existing model.Document
	if err := json.Unmarshal(raw, &existing); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	if model.IsExpired(existing.Meta) {
		txn.Rollback()
		return nil, ErrKeyNotFound
	}

	err = model.ValidateExpiration(opts.Expiration)
	if err != nil {
		txn.Rollback()
		return nil, err
	}

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

	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to replace document: %w", err)
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &doc, nil
}
