package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

var ErrRevisionMismatch = errors.New("revision mismatch")

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
		return nil, fmt.Errorf("column family '%s' does not exist", opts.ColumnFamily)
	}

	if opts.Key == "" {
		return nil, errors.New("key cannot be empty")
	}

	if opts.Value == nil {
		return nil, errors.New("value cannot be nil")
	}

	if err := model.ValidateValue(opts.Value, opts.Type); err != nil {
		return nil, fmt.Errorf("invalid value for type %s: %w", opts.Type, err)
	}

	if opts.Cas != "" {
		readOpts := grocksdb.NewDefaultReadOptions()
		readOpts.SetFillCache(false)
		defer readOpts.Destroy()

		existingDoc, err := db.Get(opts.ColumnFamily, opts.Key, readOpts)
		if err != nil {
			return nil, err
		}
		if existingDoc != nil && existingDoc.Meta.Rev != opts.Cas {
			return nil, ErrRevisionMismatch
		}
	}

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

	data, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("error marshaling document: %w", err)
	}

	err = db.TransactionDB.PutCF(opts.WriteOptions, handle, []byte(opts.Key), data)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}
