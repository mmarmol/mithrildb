package bootstrap

import (
	"log"
	"mithrildb/config"
	"mithrildb/db"
)

// InitDatabase opens RocksDB and wraps it in the mithrildb DB struct.
func InitDatabase(cfg *config.AppConfig) *db.DB {
	if cfg.RocksDB == nil {
		log.Fatal("❌ [Database.RocksDB] section is required in config.ini")
	}

	rocksdb, families, err := db.OpenRocksDBWithConfig(*cfg.RocksDB)
	if err != nil {
		log.Fatalf("Error initializing RocksDB: %v", err)
	}

	database := db.NewDB(rocksdb, families, cfg)
	return database
}
