package handlers

import (
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

func DeleteHandler(database *db.DB, defaults config.WriteOptionsConfig, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := getCfQueryParam(r)

		// Determine write options
		opts := database.DefaultWriteOptions
		override := db.HasWriteOptions(r)
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		// Call Delete with the specified column family
		if err := database.DeleteDirect(cf, key, opts); err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
