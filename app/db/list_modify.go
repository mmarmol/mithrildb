package db

import (
	"encoding/json"
	"fmt"
	"mithrildb/model"
	"time"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// PushToList appends an element to the end of a list document.
func (db *DB) PushToList(opts ListPushOptions) (interface{}, error) {
	return db.withListTransaction(opts.ListOpOptions, func(list []interface{}) ([]interface{}, interface{}, error) {
		return append(list, opts.Element), nil, nil
	})
}

// UnshiftToList inserts an element at the beginning of a list document.
func (db *DB) UnshiftToList(opts ListPushOptions) (interface{}, error) {
	return db.withListTransaction(opts.ListOpOptions, func(list []interface{}) ([]interface{}, interface{}, error) {
		return append([]interface{}{opts.Element}, list...), nil, nil
	})
}

// PopFromList removes the last element from a list document and returns it.
func (db *DB) PopFromList(opts ListOpOptions) (interface{}, error) {
	return db.withListTransaction(opts, func(list []interface{}) ([]interface{}, interface{}, error) {
		if len(list) == 0 {
			return nil, nil, ErrEmptyList
		}
		last := list[len(list)-1]
		return list[:len(list)-1], last, nil
	})
}

// ShiftFromList removes the first element from a list document and returns it.
func (db *DB) ShiftFromList(opts ListOpOptions) (interface{}, error) {
	return db.withListTransaction(opts, func(list []interface{}) ([]interface{}, interface{}, error) {
		if len(list) == 0 {
			return nil, nil, ErrEmptyList
		}
		first := list[0]
		return list[1:], first, nil
	})
}

// withListTransaction applies a list-modifying function transactionally to a list document.
func (db *DB) withListTransaction(
	opts ListOpOptions,
	modifier func([]interface{}) ([]interface{}, interface{}, error),
) (interface{}, error) {
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

	var doc model.Document
	if err := json.Unmarshal(val.Data(), &doc); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	if model.IsExpired(doc.Meta) {
		txn.Rollback()
		return nil, ErrKeyNotFound
	}

	list, ok := doc.Value.([]interface{})
	if !ok {
		txn.Rollback()
		return nil, ErrInvalidListType
	}

	newList, result, err := modifier(list)
	if err != nil {
		txn.Rollback()
		return nil, err
	}

	doc.Value = newList
	doc.Meta.Rev = uuid.NewString()
	doc.Meta.UpdatedAt = time.Now()

	if opts.Expiration != nil {
		if err := model.ValidateExpiration(*opts.Expiration); err != nil {
			txn.Rollback()
			return nil, err
		}
		doc.Meta.Expiration = *opts.Expiration
	}

	data, err := json.Marshal(doc)
	if err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to marshal document: %w", err)
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

	return result, nil
}
