package db

import (
	"mithrildb/config"
	"net/http"
	"strings"

	"github.com/linxGnu/grocksdb"
)

// DocumentReadOptions contains options for reading a single document.
type DocumentReadOptions struct {
	ColumnFamily string
	Key          string
	ReadOptions  *grocksdb.ReadOptions
}

// DocumentWriteOptions defines configurable parameters for document insertion or update.
type DocumentWriteOptions struct {
	ColumnFamily string
	Key          string
	Value        interface{}
	Cas          string
	Type         string
	Expiration   *int64
	WriteOptions *grocksdb.WriteOptions
}

// BulkWriteOptions contains parameters for inserting or replacing multiple documents.
type BulkWriteOptions struct {
	ColumnFamily string
	Documents    map[string]interface{}
	Expiration   *int64
	WriteOptions *grocksdb.WriteOptions
}

// BulkReadOptions contains parameters for reading multiple documents.
type BulkReadOptions struct {
	ColumnFamily string
	Keys         []string
	ReadOptions  *grocksdb.ReadOptions
}

// CounterIncrementOptions holds the parameters for incrementing a counter document.
type CounterIncrementOptions struct {
	ColumnFamily string
	Key          string
	Delta        int64
	Expiration   *int64
	WriteOptions *grocksdb.WriteOptions
}

// DocumentDeleteOptions contains parameters for deleting a document.
type DocumentDeleteOptions struct {
	ColumnFamily string
	Key          string
	WriteOptions *grocksdb.WriteOptions
}

// KeyListOptions defines parameters to list document keys from a column family.
type KeyListOptions struct {
	ColumnFamily string
	Prefix       string
	StartAfter   string
	Limit        int
	ReadOptions  *grocksdb.ReadOptions
}

// ListRangeOptions defines parameters for slicing a list document by index range.
type ListRangeOptions struct {
	ColumnFamily string
	Key          string
	Start        int
	End          int
	ReadOptions  *grocksdb.ReadOptions
}

// SetContainsOptions defines parameters to check if a set contains an element.
type SetContainsOptions struct {
	ColumnFamily string
	Key          string
	Element      interface{}
	ReadOptions  *grocksdb.ReadOptions
}

// ListOpOptions defines base parameters for list operations.
type ListOpOptions struct {
	ColumnFamily string
	Key          string
	WriteOptions *grocksdb.WriteOptions
	Expiration   *int64
}

// ListPushOptions extends ListOpOptions with an element to push/unshift.
type ListPushOptions struct {
	ListOpOptions
	Element interface{}
}

// SetOpOptions defines base parameters for set operations.
type SetOpOptions = ListOpOptions

func HasWriteOptions(r *http.Request) bool {
	return r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
}

func BuildWriteOptions(r *http.Request, defaults config.WriteOptionsConfig) *grocksdb.WriteOptions {
	opts := grocksdb.NewDefaultWriteOptions()

	sync := r.URL.Query().Get("sync")
	if sync == "" {
		opts.SetSync(defaults.Sync)
	} else {
		opts.SetSync(sync == "true")
	}

	wal := r.URL.Query().Get("disable_wal")
	if wal == "" {
		opts.DisableWAL(defaults.DisableWAL)
	} else {
		opts.DisableWAL(wal == "true")
	}

	slowdown := r.URL.Query().Get("no_slowdown")
	if slowdown == "" {
		opts.SetNoSlowdown(defaults.NoSlowdown)
	} else {
		opts.SetNoSlowdown(slowdown == "true")
	}

	return opts
}

func HasReadOptions(r *http.Request) bool {
	return r.URL.Query().Has("fill_cache") || r.URL.Query().Has("read_tier")
}

func BuildReadOptions(r *http.Request, defaults config.ReadOptionsConfig) *grocksdb.ReadOptions {
	opts := grocksdb.NewDefaultReadOptions()

	fill := r.URL.Query().Get("fill_cache")
	if fill == "" {
		opts.SetFillCache(defaults.FillCache)
	} else {
		opts.SetFillCache(fill == "true")
	}

	tier := r.URL.Query().Get("read_tier")
	if tier == "" {
		tier = defaults.ReadTier
	}
	if strings.ToLower(tier) == "cache-only" {
		opts.SetReadTier(ReadTierCacheOnly)
	} else {
		opts.SetReadTier(ReadTierAll)
	}

	return opts
}
