package db

import (
	"github.com/linxGnu/grocksdb"
)

func (db *DB) DeleteDirect(cf, key string, opts *grocksdb.WriteOptions) error {
	handle, ok := db.Families[cf]
	if !ok {
		return ErrInvalidColumnFamily
	}

	return db.TransactionDB.DeleteCF(opts, handle, []byte(key))
}
