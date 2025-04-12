package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

func MultiPutHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode the JSON body into the payload map
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if len(payload) == 0 {
			http.Error(w, "empty key-value map", http.StatusBadRequest)
			return
		}

		// Determine the column family (cf) to use
		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default" // If no column family is provided, use the default one
		}

		// Determine if we need to override the default write options
		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		// Call MultiPut with the specified column family and write options
		if err := database.MultiPut(cf, payload, opts); err != nil {
			http.Error(w, "multi put failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return a successful status
		w.WriteHeader(http.StatusOK)
	}
}
