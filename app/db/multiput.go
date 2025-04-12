package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

// MultiPut stores multiple key-value pairs as Documents with metadata.
func (db *DB) MultiPut(cf string, pairs map[string]string, opts *grocksdb.WriteOptions) error {
	handle, ok := db.Families[cf]
	if !ok {
		return fmt.Errorf("column family '%s' does not exist", cf)
	}

	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	now := time.Now()
	for k, v := range pairs {
		doc := model.Document{
			Key:   k,
			Value: v,
			Meta: model.Metadata{
				Rev:        uuid.NewString(),
				Type:       model.DocTypeJSON,
				UpdatedAt:  now,
				Expiration: 0,
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
