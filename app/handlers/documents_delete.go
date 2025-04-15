package handlers

import (
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

func documentDeleteHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Required: key
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
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
