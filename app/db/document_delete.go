package db

import (
	"fmt"
	"mithrildb/events"
	"mithrildb/model"

	"github.com/linxGnu/grocksdb"
)

// DeleteDocument removes a document and its TTL index (if any) in an atomic transaction.
func (db *DB) DeleteDocument(opts DocumentDeleteOptions) error {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return ErrInvalidColumnFamily
	}
	if err := model.ValidateDocumentKey(opts.Key); err != nil {
		return err
	}

	txnOpts := grocksdb.NewDefaultTransactionOptions()
	txnOpts.SetSetSnapshot(true)
	defer txnOpts.Destroy()

	txn := db.TransactionDB.TransactionBegin(opts.WriteOptions, txnOpts, nil)
	defer txn.Destroy()

	if err := txn.DeleteCF(handle, []byte(opts.Key)); err != nil {
		txn.Rollback()
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Publicar evento de eliminación sin documento ni expiración
	if err := events.PublishChangeEvent(events.ChangeEventOptions{
		Txn:       txn,
		CFName:    opts.ColumnFamily,
		Key:       opts.Key,
		Operation: events.OpDelete,
		Document:  nil,
	}); err != nil {
		txn.Rollback()
		return fmt.Errorf("failed to enqueue delete event: %w", err)
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return fmt.Errorf("failed to commit delete: %w", err)
	}

	return nil
}
