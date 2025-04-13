package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"net/http"
)

// ConfigGetHandler handles GET /config
func ConfigGetHandler(cfg config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
	}
}
