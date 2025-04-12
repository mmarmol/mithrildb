package db

import (
	"fmt"

	"github.com/linxGnu/grocksdb"
)

// DB representa una conexión a la base de datos RocksDB.
type DB struct {
	TransactionDB *grocksdb.TransactionDB
}

// RocksDBOptions contiene las opciones de configuración para RocksDB.
type RocksDBOptions struct {
	DBPath string
}

// Open inicializa la conexión a la base de datos con las opciones proporcionadas.
func Open(options RocksDBOptions) (*DB, error) {
	// Opciones generales de la DB
	opts := grocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	defer opts.Destroy()

	// Opciones específicas para TransactionDB
	txnOpts := grocksdb.NewDefaultTransactionDBOptions()
	defer txnOpts.Destroy()

	// Ahora pasás las 3 opciones necesarias
	db, err := grocksdb.OpenTransactionDb(opts, txnOpts, options.DBPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la base de datos: %w", err)
	}
	return &DB{TransactionDB: db}, nil
}

// Close cierra la conexión a la base de datos.
func (db *DB) Close() {
	db.TransactionDB.Close()
}
