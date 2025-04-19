package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// BulkPutDocuments writes multiple documents in a single atomic transaction.
// All documents must be valid and optionally share a common expiration.
func (db *DB) BulkPutDocuments(opts BulkWriteOptions) error {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return ErrInvalidColumnFamily
	}

	if opts.Expiration != nil {
		if err := model.ValidateExpiration(*opts.Expiration); err != nil {
			return err
		}
	}

	txnOpts := grocksdb.NewDefaultTransactionOptions()
	txnOpts.SetSetSnapshot(true)
	defer txnOpts.Destroy()

	txn := db.TransactionDB.TransactionBegin(opts.WriteOptions, txnOpts, nil)
	defer txn.Destroy()

	now := time.Now()

	for k, rawValue := range opts.Documents {
		if err := model.ValidateDocumentKey(k); err != nil {
			txn.Rollback()
			return fmt.Errorf("invalid key '%s': %w", k, err)
		}
		if err := model.ValidateValue(rawValue, model.DocTypeJSON); err != nil {
			txn.Rollback()
			return fmt.Errorf("invalid value for key '%s': %w", k, err)
		}

		exp := int64(0)
		if opts.Expiration != nil {
			exp = *opts.Expiration
		}

		doc := model.Document{
			Key:   k,
			Value: rawValue,
			Meta: model.Metadata{
				Rev:        uuid.NewString(),
				Type:       model.DocTypeJSON,
				UpdatedAt:  now,
				Expiration: exp,
			},
		}

		data, err := json.Marshal(doc)
		if err != nil {
			txn.Rollback()
			return fmt.Errorf("failed to encode document for key '%s': %w", k, err)
		}

		if err := txn.PutCF(handle, []byte(k), data); err != nil {
			txn.Rollback()
			return fmt.Errorf("failed to write document for key '%s': %w", k, err)
		}

		if opts.Expiration != nil {
			if err := db.ReplaceTTLInTxn(txn, opts.ColumnFamily, k, *opts.Expiration); err != nil {
				txn.Rollback()
				return fmt.Errorf("failed to write TTL for key '%s': %w", k, err)
			}
		}
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return fmt.Errorf("failed to commit BulkPutDocuments: %w", err)
	}

	return nil
}
