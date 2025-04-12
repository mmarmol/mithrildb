package db

import "github.com/linxGnu/grocksdb"

// Put inserta o actualiza un valor asociado a una clave en la base de datos dentro de una transacción.
func (db *DB) PutDirect(key string, value string) error {
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()

	return db.TransactionDB.Put(wo, []byte(key), []byte(value))
}
