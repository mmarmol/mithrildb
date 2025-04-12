package db

import (
	"mithrildb/config"
	"net/http"
	"strings"

	"github.com/linxGnu/grocksdb"
)

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
