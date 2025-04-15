package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/metrics"
	"net/http"
	"time"
)

// metricsHandler provides system and database metrics.
//
// @Summary      Retrieve metrics
// @Description  Returns server-level and RocksDB-specific metrics, including memory usage, uptime, and database statistics.
// @Tags         monitoring
// @Produce      json
// @Success      200  {object}  metrics.FullMetrics
// @Failure      500  {object}  handlers.ErrorResponse  "Failed to collect metrics"
// @Router       /metrics [get]
func metricsHandler(database *db.DB, cfg *config.AppConfig, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server, err := metrics.GetServerMetrics(cfg, startTime)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "error reading system metrics: "+err.Error())
			return
		}

		rocks := metrics.GetRocksDBMetrics(database.TransactionDB)

		result := metrics.FullMetrics{
			Server:  server,
			RocksDB: rocks,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
