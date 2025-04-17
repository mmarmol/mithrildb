package db

import (
	"fmt"
	"mithrildb/model"

	"github.com/linxGnu/grocksdb"
)

func (db *DB) DeleteDirect(cfName, key string, opts *grocksdb.WriteOptions) error {
	handle, ok := db.Families[cfName]
	if !ok {
		return ErrInvalidColumnFamily
	}
	if err := model.ValidateDocumentKey(key); err != nil {
		return err
	}

	txnOpts := grocksdb.NewDefaultTransactionOptions()
	txnOpts.SetSetSnapshot(true)
	defer txnOpts.Destroy()

	txn := db.TransactionDB.TransactionBegin(opts, txnOpts, nil)
	defer txn.Destroy()

	if err := txn.DeleteCF(handle, []byte(key)); err != nil {
		txn.Rollback()
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if err := db.ClearAllTTLInTxn(txn, cfName, key); err != nil {
		txn.Rollback()
		return fmt.Errorf("failed to remove TTL index: %w", err)
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return fmt.Errorf("failed to commit delete: %w", err)
	}

	return nil
}
