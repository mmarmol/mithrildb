package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"strconv"
)

// ListKeysHandler handles GET /keys
// Supports optional parameters: cf, prefix, start_after, limit, fill_cache, read_tier
func ListKeysHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default"
		}

		prefix := r.URL.Query().Get("prefix")
		startAfter := r.URL.Query().Get("start_after")
		limitStr := r.URL.Query().Get("limit")

		limit := 100
		if limitStr != "" {
			if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
				limit = val
			}
		}

		// Read options
		opts := database.DefaultReadOptions
		override := r.URL.Query().Has("fill_cache") || r.URL.Query().Has("read_tier")
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		keys, err := database.ListKeys(cf, prefix, startAfter, limit, opts)
		if err != nil {
			http.Error(w, "failed to list keys: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(keys)
	}
}
