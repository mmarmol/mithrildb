package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// updateDocumentIfExists executes a transactional update over an existing document.
func (db *DB) updateDocumentIfExists(
	opts PutOptions,
	modify func(doc *model.Document) error,
) (*model.Document, error) {
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

	var existing model.Document
	if err := json.Unmarshal(val.Data(), &existing); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	if model.IsExpired(existing.Meta) {
		txn.Rollback()
		return nil, ErrKeyNotFound
	}

	if err := modify(&existing); err != nil {
		txn.Rollback()
		return nil, err
	}

	data, err := json.Marshal(existing)
	if err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to serialize document: %w", err)
	}

	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to write document: %w", err)
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &existing, nil
}

// Replace stores a Document only if the key already exists.
func (db *DB) Replace(opts PutOptions) (*model.Document, error) {
	return db.updateDocumentIfExists(opts, func(doc *model.Document) error {
		if opts.Value == nil {
			return ErrNilValue
		}
		if err := model.ValidateValue(opts.Value, opts.Type); err != nil {
			return fmt.Errorf("invalid value for type %s: %w", opts.Type, err)
		}
		if err := model.ValidateExpiration(opts.Expiration); err != nil {
			return err
		}
		doc.Value = opts.Value
		doc.Meta.Type = opts.Type
		doc.Meta.Expiration = opts.Expiration
		doc.Meta.UpdatedAt = time.Now().UTC()
		doc.Meta.Rev = uuid.NewString()
		return nil
	})
}

// Touch updates the expiration of a document without modifying its content.
func (db *DB) Touch(opts PutOptions) (*model.Document, error) {
	return db.updateDocumentIfExists(opts, func(doc *model.Document) error {
		if err := model.ValidateExpiration(opts.Expiration); err != nil {
			return err
		}
		doc.Meta.Expiration = opts.Expiration
		doc.Meta.UpdatedAt = time.Now().UTC()
		doc.Meta.Rev = uuid.NewString()
		return nil
	})
}
