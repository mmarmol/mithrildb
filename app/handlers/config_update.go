// handlers/config_update.go
package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// ConfigUpdateHandler handles POST /config/update
func ConfigUpdateHandler(cfg *config.AppConfig, dbInstance *db.DB, configPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		result, err := config.UpdateConfigFromMap(cfg, configPath, req)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not update configuration")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
