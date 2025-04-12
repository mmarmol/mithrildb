package db

import (
	"fmt"

	"github.com/linxGnu/grocksdb"
)

func (db *DB) DeleteDirect(cf, key string, opts *grocksdb.WriteOptions) error {
	handle, ok := db.Families[cf]
	if !ok {
		return fmt.Errorf("column family '%s' does not exist", cf)
	}

	return db.TransactionDB.DeleteCF(opts, handle, []byte(key))
}
