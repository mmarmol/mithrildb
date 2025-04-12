package db

import "github.com/linxGnu/grocksdb"

// MultiPut stores multiple key-value pairs atomically using WriteBatch.
func (db *DB) MultiPut(pairs map[string]string, opts *grocksdb.WriteOptions) error {
	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	for k, v := range pairs {
		batch.Put([]byte(k), []byte(v))
	}

	return db.TransactionDB.Write(opts, batch)
}
