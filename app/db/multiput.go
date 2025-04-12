package db

import (
	"fmt"

	"github.com/linxGnu/grocksdb"
)

func (db *DB) MultiPut(cf string, pairs map[string]string, opts *grocksdb.WriteOptions) error {
	handle, ok := db.Families[cf]
	if !ok {
		return fmt.Errorf("column family '%s' does not exist", cf)
	}

	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	for k, v := range pairs {
		batch.PutCF(handle, []byte(k), []byte(v))
	}

	return db.TransactionDB.Write(opts, batch)
}
