package db

import (
	"mithrildb/model"

	"github.com/linxGnu/grocksdb"
)

func (db *DB) DeleteDirect(cf, key string, opts *grocksdb.WriteOptions) error {
	handle, ok := db.Families[cf]
	if !ok {
		return ErrInvalidColumnFamily
	}
	err := model.ValidateDocumentKey(key)
	if err != nil {
		return err
	}

	return db.TransactionDB.DeleteCF(opts, handle, []byte(key))
}
