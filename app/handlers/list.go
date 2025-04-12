package handlers

import (
	"encoding/json"
	"mithrildb/db"
	"net/http"
	"strconv"
)

func ListHandler(database *db.DB) http.HandlerFunc {
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

		opts := database.DefaultReadOptions

		keys, err := database.ListKeys(cf, prefix, startAfter, limit, opts)
		if err != nil {
			http.Error(w, "failed to list keys: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(keys)
	}
}
