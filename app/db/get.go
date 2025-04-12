package db

import "github.com/linxGnu/grocksdb"

// Get recupera un valor utilizando una clave de la base de datos.
func (db *DB) Get(key string) (string, error) {
	readOpts := grocksdb.NewDefaultReadOptions()
	defer readOpts.Destroy() // Limpia las opciones de lectura

	value, err := db.TransactionDB.Get(readOpts, []byte(key))
	if err != nil {
		return "", err
	}
	defer value.Free() // ⚠️ Muy importante liberar la Slice

	if value.Size() == 0 {
		return "", nil // Clave inexistente o valor vacío
	}

	return string(value.Data()), nil
}
