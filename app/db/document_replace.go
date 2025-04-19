package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// updateIfExists executes a transactional update on an existing document.
func (db *DB) updateIfExists(
	opts DocumentWriteOptions,
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

	return &existing, nil
}

// ReplaceDocument overwrites a document only if the key already exists.
func (db *DB) ReplaceDocument(opts DocumentWriteOptions) (*model.Document, error) {
	return db.updateIfExists(opts, func(doc *model.Document) error {
		if opts.Value == nil {
			return ErrNilValue
		}
		if err := model.ValidateValue(opts.Value, opts.Type); err != nil {
			return fmt.Errorf("invalid value for type %s: %w", opts.Type, err)
		}
		if opts.Expiration != nil {
			if err := model.ValidateExpiration(*opts.Expiration); err != nil {
				return err
			}
			doc.Meta.Expiration = *opts.Expiration
		}
		doc.Value = opts.Value
		doc.Meta.Type = opts.Type
		doc.Meta.UpdatedAt = time.Now().UTC()
		doc.Meta.Rev = uuid.NewString()
		return nil
	})
}

// TouchDocument updates only the expiration timestamp of an existing document.
func (db *DB) TouchDocument(opts DocumentWriteOptions) (*model.Document, error) {
	return db.updateIfExists(opts, func(doc *model.Document) error {
		if opts.Expiration == nil {
			return model.ErrInvalidExpiration
		}
		if err := model.ValidateExpiration(*opts.Expiration); err != nil {
			return err
		}
		doc.Meta.Expiration = *opts.Expiration
		doc.Meta.UpdatedAt = time.Now().UTC()
		doc.Meta.Rev = uuid.NewString()
		return nil
	})
}
