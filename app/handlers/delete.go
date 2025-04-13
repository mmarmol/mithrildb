package handlers

import (
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

func DeleteHandler(database *db.DB, defaults config.WriteOptionsConfig, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default" // Default column family
		}

		// Determine write options
		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		// Call Delete with the specified column family
		if err := database.DeleteDirect(cf, key, opts); err != nil {
			if err == db.ErrInvalidColumnFamily {
				respondWithErrInvalidColumnFamily(w, cf)
				return
			}
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
