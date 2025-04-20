package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/events"
	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// PutDocument stores or updates a document with optional CAS and expiration.
func (db *DB) PutDocument(opts DocumentWriteOptions) (*model.Document, error) {
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
	txnOpts.SetLockTimeout(1000)
	defer txnOpts.Destroy()

	txn := db.TransactionDB.TransactionBegin(opts.WriteOptions, txnOpts, nil)
	defer txn.Destroy()

	var prevMeta *model.Metadata
	if opts.Cas != "" {
		readOpts := grocksdb.NewDefaultReadOptions()
		readOpts.SetFillCache(false)
		defer readOpts.Destroy()

		val, err := txn.GetWithCF(readOpts, handle, []byte(opts.Key))
		if err != nil {
			txn.Rollback()
			return nil, err
		}
		defer val.Free()

		if val.Size() > 0 {
			var existing model.Document
			if err := json.Unmarshal(val.Data(), &existing); err != nil {
				txn.Rollback()
				return nil, fmt.Errorf("failed to unmarshal existing document: %w", err)
			}
			if existing.Meta.Rev != opts.Cas {
				txn.Rollback()
				return nil, ErrRevisionMismatch
			}
			metaCopy := existing.Meta
			prevMeta = &metaCopy
		}
	}

	now := time.Now()
	exp := int64(0)
	if opts.Expiration != nil {
		exp = *opts.Expiration
	}

	doc := model.Document{
		Key:   opts.Key,
		Value: opts.Value,
		Meta: model.Metadata{
			Rev:        uuid.NewString(),
			Type:       opts.Type,
			UpdatedAt:  now,
			Expiration: exp,
		},
	}

	data, err := json.Marshal(doc)
	if err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("error marshaling document: %w", err)
	}

	if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to put document: %w", err)
	}

	if err := events.PublishChangeEvent(events.ChangeEventOptions{
		Txn:                txn,
		CFName:             opts.ColumnFamily,
		Key:                opts.Key,
		Document:           &doc,
		Operation:          events.OpPut,
		PreviousMeta:       prevMeta,
		ExplicitExpiration: opts.Expiration,
	}); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to enqueue change event: %w", err)
	}

	if err := txn.Commit(); err != nil {
		txn.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &doc, nil
}
