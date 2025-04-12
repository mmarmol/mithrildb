package db

import (
	"fmt"

	"github.com/linxGnu/grocksdb"
)

func (db *DB) Get(cf, key string, opts *grocksdb.ReadOptions) (string, error) {
	handle, ok := db.Families[cf]
	if !ok {
		return "", fmt.Errorf("column family '%s' does not exist", cf)
	}

	value, err := db.TransactionDB.GetCF(opts, handle, []byte(key))
	if err != nil {
		return "", err
	}
	defer value.Free()

	if value.Size() == 0 {
		return "", nil
	}

	return string(value.Data()), nil
}
