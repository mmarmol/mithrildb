package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"net/http"
)

// configGetHandler handles GET /config
func configGetHandler(cfg config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
	}
}
