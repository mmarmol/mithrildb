package db

import (
	"encoding/json"
	"fmt"
	"mithrildb/model"
	"time"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// ListOpOptions define los valores necesarios para modificar un documento tipo lista
type ListOpOptions struct {
	ColumnFamily string
	Key          string
	WriteOptions *grocksdb.WriteOptions
	Expiration   int64
}

// withListTransaction permite modificar un documento lista de forma transaccional y reutilizable
func (db *DB) withListTransaction(opts ListOpOptions, modifier func([]interface{}) ([]interface{}, interface{}, error)) (interface{}, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
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
	doc.Meta.Expiration = opts.Expiration

	data, err := json.Marshal(doc)
	if err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to write document: %w", err)
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// ListPush agrega un elemento al final de la lista
type ListPushOptions struct {
	ListOpOptions
	Element interface{}
}

func (db *DB) ListPush(opts ListPushOptions) (interface{}, error) {
	return db.withListTransaction(opts.ListOpOptions, func(list []interface{}) ([]interface{}, interface{}, error) {
		return append(list, opts.Element), nil, nil
	})
}

// ListUnshift agrega un elemento al principio de la lista
func (db *DB) ListUnshift(opts ListPushOptions) (interface{}, error) {
	return db.withListTransaction(opts.ListOpOptions, func(list []interface{}) ([]interface{}, interface{}, error) {
		return append([]interface{}{opts.Element}, list...), nil, nil
	})
}

// ListPop elimina y devuelve el último elemento
func (db *DB) ListPop(opts ListOpOptions) (interface{}, error) {
	return db.withListTransaction(opts, func(list []interface{}) ([]interface{}, interface{}, error) {
		if len(list) == 0 {
			return nil, nil, ErrEmptyList
		}
		last := list[len(list)-1]
		return list[:len(list)-1], last, nil
	})
}

// ListShift elimina y devuelve el primer elemento
func (db *DB) ListShift(opts ListOpOptions) (interface{}, error) {
	return db.withListTransaction(opts, func(list []interface{}) ([]interface{}, interface{}, error) {
		if len(list) == 0 {
			return nil, nil, ErrEmptyList
		}
		first := list[0]
		return list[1:], first, nil
	})
}
