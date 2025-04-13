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
		cf := getCfQueryParam(r)

		opts := database.DefaultReadOptions
		override := db.HasReadOptions(r)
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		doc, err := database.Get(cf, key, opts)
		if err != nil {
			mapAndRespondWithError(w, err)
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
