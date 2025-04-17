package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

func (db *DB) MultiPut(cf string, pairs map[string]interface{}, expiration *int64, opts *grocksdb.WriteOptions) error {
	handle, ok := db.Families[cf]
	if !ok {
		return ErrInvalidColumnFamily
	}

	if expiration != nil {
		if err := model.ValidateExpiration(*expiration); err != nil {
			return err
		}
	}

	txnOpts := grocksdb.NewDefaultTransactionOptions()
	txnOpts.SetSetSnapshot(true)
	defer txnOpts.Destroy()

	txn := db.TransactionDB.TransactionBegin(opts, txnOpts, nil)
	defer txn.Destroy()

	now := time.Now()

	for k, rawValue := range pairs {
		if err := model.ValidateDocumentKey(k); err != nil {
			txn.Rollback()
			return fmt.Errorf("invalid key '%s': %w", k, err)
		}
		if err := model.ValidateValue(rawValue, model.DocTypeJSON); err != nil {
			txn.Rollback()
			return fmt.Errorf("invalid value for key '%s': %w", k, err)
		}

		exp := int64(0)
		if expiration != nil {
			exp = *expiration
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

		if expiration != nil {
			if err := db.ReplaceTTLInTxn(txn, cf, k, *expiration); err != nil {
				txn.Rollback()
				return fmt.Errorf("failed to write TTL for key '%s': %w", k, err)
			}
		}
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return fmt.Errorf("failed to commit MultiPut: %w", err)
	}

	return nil
}
