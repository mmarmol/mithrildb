package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"strconv"
)

// ListDocumentsHandler handles GET /documents with optional prefix, pagination and read options.
func ListDocumentsHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
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

		// Read options (use default for now)
		opts := database.DefaultReadOptions
		override := r.URL.Query().Has("fill_cache") || r.URL.Query().Has("read_tier")
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		// First list keys
		keys, err := database.ListKeys(cf, prefix, startAfter, limit, opts)
		if err != nil {
			http.Error(w, "failed to list keys: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Then fetch documents for those keys
		docs, err := database.MultiGet(cf, keys, opts)
		if err != nil {
			http.Error(w, "failed to retrieve documents: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return map of key => document
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(docs)
	}
}
