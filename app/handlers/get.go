package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// GetHandler retrieves a full document including metadata.
func GetHandler(database *db.DB, defaults config.ReadOptionsConfig, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default"
		}

		opts := database.DefaultReadOptions
		override := r.URL.Query().Has("fill_cache") || r.URL.Query().Has("read_tier")
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		doc, err := database.Get(cf, key, opts)
		if err != nil {
			if err == db.ErrInvalidColumnFamily {
				respondWithErrInvalidColumnFamily(w, cf)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if doc == nil {
			respondWithError(w, http.StatusNotFound, "Key not found")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	}
}
