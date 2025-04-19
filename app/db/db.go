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

// DB represents the main RocksDB interface with configured options and column families.
type DB struct {
	TransactionDB       *grocksdb.TransactionDB
	DefaultReadOptions  *grocksdb.ReadOptions
	DefaultWriteOptions *grocksdb.WriteOptions
	Families            map[string]*grocksdb.ColumnFamilyHandle
	mu                  sync.Mutex
}

// NewDB creates a DB instance with default read and write options based on the application config.
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

// Close releases resources associated with the database.
func (db *DB) Close() {
	db.DefaultReadOptions.Destroy()
	db.DefaultWriteOptions.Destroy()
	db.TransactionDB.Close()
}

// OpenRocksDBWithConfig opens or creates a RocksDB instance with support for multiple column families.
func OpenRocksDBWithConfig(cfg config.RocksDBConfig) (*grocksdb.TransactionDB, map[string]*grocksdb.ColumnFamilyHandle, error) {
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

	// Block-based table and cache
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	cache := grocksdb.NewLRUCache(uint64(cfg.BlockCacheSize))
	bbto.SetBlockCache(cache)
	opts.SetBlockBasedTableFactory(bbto)

	// Compression
	opts.SetCompression(parseCompressionType(cfg.CompressionType, cfg.EnableCompression))

	txnOpts := grocksdb.NewDefaultTransactionDBOptions()
	dbPath := cfg.DBPath

	var cfNames []string
	if _, err := os.Stat(dbPath); err == nil {
		cfNames, err = grocksdb.ListColumnFamilies(opts, dbPath)
		if err != nil || len(cfNames) == 0 {
			fmt.Println("⚠️  No column families found, initializing with 'default'")
			cfNames = []string{"default"}
		}
	} else {
		cfNames = []string{"default"}
	}

	cfOpts := make([]*grocksdb.Options, len(cfNames))
	for i := range cfNames {
		cfOpts[i] = opts
	}

	db, cfHandles, err := grocksdb.OpenTransactionDbColumnFamilies(opts, txnOpts, dbPath, cfNames, cfOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open DB with column families: %w", err)
	}

	handles := make(map[string]*grocksdb.ColumnFamilyHandle, len(cfNames))
	for i, name := range cfNames {
		handles[strings.ToLower(name)] = cfHandles[i]
	}

	return db, handles, nil
}

// parseCompressionType converts a string compression type to the corresponding RocksDB constant.
func parseCompressionType(name string, enabled bool) grocksdb.CompressionType {
	if !enabled {
		return grocksdb.NoCompression
	}
	switch strings.ToLower(name) {
	case "snappy":
		return grocksdb.SnappyCompression
	case "zstd":
		return grocksdb.ZSTDCompression
	case "lz4":
		return grocksdb.LZ4Compression
	case "none":
		return grocksdb.NoCompression
	default:
		fmt.Printf("⚠️  Unknown compression type: %s, defaulting to snappy\n", name)
		return grocksdb.SnappyCompression
	}
}
