// handlers/config_update.go
package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"net/http"
)

// configUpdateHandler updates RocksDB configuration parameters at runtime.
//
// @Summary      Update configuration
// @Description  Updates system configuration parameters. Some changes are applied live, while others require a restart.
// @Tags         config
// @Accept       json
// @Produce      json
// @Param        body  body      map[string]interface{}  true  "Configuration parameters to update"
// @Success      200   {object}  config.UpdateResult
// @Failure      400   {object}  handlers.ErrorResponse  "Invalid JSON body"
// @Failure      500   {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /config/update [put]
func configUpdateHandler(cfg *config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		result, err := config.UpdateConfigFromMap(cfg, req)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not update configuration")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
