package db

import "github.com/linxGnu/grocksdb"

// Delete elimina un valor utilizando una clave de la base de datos dentro de una transacción.
func (db *DB) DeleteDirect(key string) error {
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()

	return db.TransactionDB.Delete(wo, []byte(key))
}
