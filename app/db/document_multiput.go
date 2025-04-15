package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

func (db *DB) MultiPut(cf string, pairs map[string]interface{}, expiration int64, opts *grocksdb.WriteOptions) error {
	handle, ok := db.Families[cf]
	if !ok {
		return ErrInvalidColumnFamily
	}

	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	now := time.Now()

	for k, rawValue := range pairs {
		// Validate value type (defaults to json for now)
		err := model.ValidateDocumentKey(k)
		if err != nil {
			return err
		}
		if err := model.ValidateValue(rawValue, model.DocTypeJSON); err != nil {
			return fmt.Errorf("invalid value for key '%s': %w", k, err)
		}

		doc := model.Document{
			Key:   k,
			Value: rawValue,
			Meta: model.Metadata{
				Rev:        uuid.NewString(),
				Type:       model.DocTypeJSON, // All are json for now
				UpdatedAt:  now,
				Expiration: expiration,
			},
		}

		data, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to encode document for key '%s': %w", k, err)
		}

		batch.PutCF(handle, []byte(k), data)
	}

	return db.TransactionDB.Write(opts, batch)
}
