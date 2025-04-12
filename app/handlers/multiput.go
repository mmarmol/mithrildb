package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

func MultiPutHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if len(payload) == 0 {
			http.Error(w, "empty key-value map", http.StatusBadRequest)
			return
		}

		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		if err := database.MultiPut(payload, opts); err != nil {
			http.Error(w, "multi put failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
