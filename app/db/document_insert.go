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
		return nil, fmt.Errorf("failed to serialize document: %w", err)
	}

	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to insert document: %w", err)
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
