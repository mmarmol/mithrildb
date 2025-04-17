package db

import (
	"fmt"
	"mithrildb/config"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/linxGnu/grocksdb"
)

// DB representa una conexión a la base de datos RocksDB.
type DB struct {
	TransactionDB       *grocksdb.TransactionDB
	DefaultReadOptions  *grocksdb.ReadOptions
	DefaultWriteOptions *grocksdb.WriteOptions
	Families            map[string]*grocksdb.ColumnFamilyHandle
	mu                  sync.Mutex
}

// RocksDBOptions contiene las opciones de configuración para RocksDB.
type RocksDBOptions struct {
	DBPath string
}

func NewDB(rocks *grocksdb.TransactionDB, families map[string]*grocksdb.ColumnFamilyHandle, cfg config.AppConfig) *DB {
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
		Families:            families,
		DefaultReadOptions:  readOpts,
		DefaultWriteOptions: writeOpts,
	}
}

func (db *DB) Close() {
	db.DefaultReadOptions.Destroy()
	db.DefaultWriteOptions.Destroy()
	db.TransactionDB.Close()
}

// NewRocksDBFromConfig abre o crea una instancia de RocksDB con soporte para múltiples column families.
func NewRocksDBFromConfig(cfg config.RocksDBConfig) (*grocksdb.TransactionDB, map[string]*grocksdb.ColumnFamilyHandle, error) {
	opts := grocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(cfg.CreateIfMissing)
	opts.SetWriteBufferSize(uint64(cfg.WriteBufferSize))
	opts.SetMaxWriteBufferNumber(cfg.MaxWriteBufferNum)
	opts.SetMaxOpenFiles(cfg.MaxOpenFiles)
	dur, err := time.ParseDuration(cfg.StatsDumpPeriod)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid stats_dump_period: %w", err)
	}
	opts.SetStatsDumpPeriodSec(uint(dur.Seconds()))

	// Configuración de tabla basada en bloques y cache
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
	dbPath := cfg.DBPath

	var cfNames []string
	if _, err := os.Stat(dbPath); err == nil {
		cfNames, err = grocksdb.ListColumnFamilies(opts, dbPath)
		if err != nil {
			// Esto indica que el directorio existe, pero no hay base aún
			fmt.Printf("⚠️  No se encontraron familias, inicializando con 'default'")
			cfNames = []string{"default"}
		} else if len(cfNames) == 0 {
			cfNames = []string{"default"}
		}
	} else {
		// Directorio no existe → crear base nueva
		cfNames = []string{"default"}
	}

	cfOpts := make([]*grocksdb.Options, len(cfNames))
	for i := range cfNames {
		// Reutilizamos las mismas opciones para cada CF
		cfOpts[i] = opts
	}

	db, cfHandles, err := grocksdb.OpenTransactionDbColumnFamilies(
		opts, txnOpts, dbPath, cfNames, cfOpts,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open DB with column families: %w", err)
	}

	handles := make(map[string]*grocksdb.ColumnFamilyHandle, len(cfNames))
	for i, name := range cfNames {
		handles[strings.ToLower(name)] = cfHandles[i]
	}

	return db, handles, nil
}
