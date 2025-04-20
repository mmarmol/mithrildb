package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/expiration"
	"mithrildb/metrics"
	"net/http"
	"time"
)

// metricsHandler provides a snapshot of internal and system metrics.
//
// @Summary      Retrieve internal metrics
// @Description  Returns server-level, RocksDB and expiration system metrics including memory usage, uptime, compaction stats, and TTL cleanup activity.
// @Tags         monitoring
// @Produce      json
// @Success      200  {object}  metrics.FullMetrics "Detailed metrics of the server, database and expiration subsystem"
// @Failure      500  {object}  handlers.ErrorResponse "Failed to collect or serialize metrics"
// @Router       /metrics [get]
func metricsHandler(database *db.DB, expirer *expiration.Service, cfg *config.AppConfig, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server, err := metrics.GetServerMetrics(cfg, startTime)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "error reading system metrics: "+err.Error())
			return
		}

		result := metrics.FullMetrics{
			Server:     server,
			RocksDB:    metrics.GetRocksDBMetrics(database.TransactionDB),
			Expiration: metrics.GetExpirationMetrics(expirer),
			Events:     metrics.GetEventSystemMetrics(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
