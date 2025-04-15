package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// PutOptions defines configurable parameters for a document insert or update.
type PutOptions struct {
	ColumnFamily string                 // Target column family
	Key          string                 // Document key
	Value        interface{}            // Document value (string, map, etc.)
	Cas          string                 // Optional revision for optimistic locking
	Type         string                 // Document type (json, counter, etc.)
	Expiration   int64                  // Optional TTL (Unix timestamp)
	WriteOptions *grocksdb.WriteOptions // RocksDB write options
}

// Put stores a Document using the provided options, handling revision control.
func (db *DB) PutWithOptions(opts PutOptions) (*model.Document, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	err := model.ValidateDocumentKey(opts.Key)
	if err != nil {
		return nil, err
	}

	if opts.Value == nil {
		return nil, ErrNilValue
	}

	if err := model.ValidateValue(opts.Value, opts.Type); err != nil {
		return nil, fmt.Errorf("invalid value for type %s: %w", opts.Type, err)
	}

	// Transaction options only if CAS is provided
	var txn *grocksdb.Transaction
	var writeOpts *grocksdb.WriteOptions
	var txnOpts *grocksdb.TransactionOptions

	// Prepare write options
	writeOpts = opts.WriteOptions // Use the write options passed in opts

	// If CAS is provided, handle it within a transaction
	if opts.Cas != "" {
		// Set up transaction options
		txnOpts = grocksdb.NewDefaultTransactionOptions()
		txnOpts.SetSetSnapshot(true) // Take a snapshot to avoid reading inconsistencies
		txnOpts.SetLockTimeout(1000) // Set lock timeout (e.g., 1 second)
		defer txnOpts.Destroy()

		// Start a new transaction
		txn = db.TransactionDB.TransactionBegin(writeOpts, txnOpts, nil)
		defer txn.Destroy()

		// Check if the document exists and its revision matches the CAS
		readOpts := grocksdb.NewDefaultReadOptions()
		readOpts.SetFillCache(false)
		defer readOpts.Destroy()

		val, err := txn.GetWithCF(readOpts, handle, []byte(opts.Key))
		if err != nil {
			txn.Rollback()
			return nil, err
		}
		defer val.Free()

		// Only check CAS if value exists
		if val.Size() > 0 {
			var existingDoc model.Document
			if err := json.Unmarshal(val.Data(), &existingDoc); err != nil {
				txn.Rollback()
				return nil, fmt.Errorf("failed to unmarshal existing document: %w", err)
			}
			if existingDoc.Meta.Rev != opts.Cas {
				txn.Rollback()
				return nil, ErrRevisionMismatch
			}
		}
	} else {
		// No CAS, no need for a transaction, use direct Put
		txn = nil
	}

	// Prepare the new document
	now := time.Now()
	doc := model.Document{
		Key:   opts.Key,
		Value: opts.Value,
		Meta: model.Metadata{
			Rev:        uuid.NewString(),
			Type:       opts.Type,
			UpdatedAt:  now,
			Expiration: opts.Expiration,
		},
	}

	// Serialize the document into JSON
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("error marshaling document: %w", err)
	}

	// If using transaction, do the Put in the transaction
	if txn != nil {
		if err := txn.PutCF(handle, []byte(opts.Key), data); err != nil {
			// If put fails, rollback the transaction
			txn.Rollback()
			return nil, fmt.Errorf("failed to put document in transaction: %w", err)
		}
		// Commit transaction after successful put
		if err := txn.Commit(); err != nil {
			txn.Rollback() // Ensure rollback on commit failure
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
	} else {
		// If no transaction, do the Put directly
		err = db.TransactionDB.PutCF(writeOpts, handle, []byte(opts.Key), data)
		if err != nil {
			return nil, fmt.Errorf("failed to put document: %w", err)
		}
	}

	// Return the stored document
	return &doc, nil
}
