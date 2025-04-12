package db

import (
	"fmt"
	"mithrildb/config"

	"github.com/linxGnu/grocksdb"
)

// DB representa una conexión a la base de datos RocksDB.
type DB struct {
	TransactionDB       *grocksdb.TransactionDB
	DefaultReadOptions  *grocksdb.ReadOptions
	DefaultWriteOptions *grocksdb.WriteOptions
}

// RocksDBOptions contiene las opciones de configuración para RocksDB.
type RocksDBOptions struct {
	DBPath string
}

func NewDB(rocks *grocksdb.TransactionDB, cfg config.AppConfig) *DB {
	readOpts := grocksdb.NewDefaultReadOptions()
	readOpts.SetFillCache(cfg.ReadDefaults.FillCache)
	if cfg.ReadDefaults.ReadTier == "cache-only" {
		readOpts.SetReadTier(grocksdb.ReadTier(ReadTierCacheOnly))
	} else {
		readOpts.SetReadTier(grocksdb.ReadTier(ReadTierAll))
	}

	writeOpts := grocksdb.NewDefaultWriteOptions()
	writeOpts.SetSync(cfg.WriteDefaults.Sync)
	writeOpts.DisableWAL(cfg.WriteDefaults.DisableWAL)
	writeOpts.SetNoSlowdown(cfg.WriteDefaults.NoSlowdown)

	return &DB{
		TransactionDB:       rocks,
		DefaultReadOptions:  readOpts,
		DefaultWriteOptions: writeOpts,
	}
}

func (db *DB) Close() {
	db.DefaultReadOptions.Destroy()
	db.DefaultWriteOptions.Destroy()
	db.TransactionDB.Close()
}

func NewRocksDBFromConfig(cfg *config.RocksDBConfig) (*grocksdb.TransactionDB, error) {
	opts := grocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(cfg.CreateIfMissing)
	opts.SetWriteBufferSize(uint64(cfg.WriteBufferSize))
	opts.SetMaxWriteBufferNumber(cfg.MaxWriteBufferNum)
	opts.SetMaxOpenFiles(cfg.MaxOpenFiles)
	opts.SetStatsDumpPeriodSec(uint(cfg.StatsDumpPeriod.Seconds()))

	// Block-based table config con cache
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	cache := grocksdb.NewLRUCache(uint64(cfg.BlockCacheSize))
	bbto.SetBlockCache(cache)
	opts.SetBlockBasedTableFactory(bbto)

	// Compresión
	if cfg.EnableCompression {
		switch cfg.CompressionType {
		case "snappy":
			opts.SetCompression(grocksdb.SnappyCompression)
		case "zstd":
			opts.SetCompression(grocksdb.ZSTDCompression)
		case "lz4":
			opts.SetCompression(grocksdb.LZ4Compression)
		case "none":
			opts.SetCompression(grocksdb.NoCompression)
		default:
			fmt.Printf("⚠️  Tipo de compresión no reconocido: %s, usando snappy\n", cfg.CompressionType)
			opts.SetCompression(grocksdb.SnappyCompression)
		}
	} else {
		opts.SetCompression(grocksdb.NoCompression)
	}

	txnOpts := grocksdb.NewDefaultTransactionDBOptions()

	db, err := grocksdb.OpenTransactionDb(opts, txnOpts, cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la base de datos RocksDB: %w", err)
	}

	return db, nil
}
