package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"net/http"
)

// configGetHandler handles GET /config
// @Summary Get server configuration
// @Description Returns the current server configuration used by the database
// @Tags config
// @Accept json
// @Produce json
// @Success 200 {object} config.AppConfig
// @Router /config [get]
func configGetHandler(cfg *config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
	}
}
