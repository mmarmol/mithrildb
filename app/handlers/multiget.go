package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// multiGetRequest defines the structure of the JSON body expected in a MultiGet request.
type multiGetRequest struct {
	Keys []string `json:"keys"`
}

// MultiGetHandler handles POST /multiget
// It receives a list of keys and returns a map of key to full document (with metadata).
// Missing keys will be returned as null.
func MultiGetHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req multiGetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if len(req.Keys) == 0 {
			http.Error(w, "empty key list", http.StatusBadRequest)
			return
		}

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

		// Perform the multi-get operation and return document objects (not just values)
		result, err := database.MultiGet(cf, req.Keys, opts)
		if err != nil {
			http.Error(w, "error reading from database: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
