package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/metrics"
	"net/http"
	"time"
)

func MetricsHandler(database *db.DB, cfg config.AppConfig, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server, err := metrics.GetServerMetrics(cfg, startTime)
		if err != nil {
			http.Error(w, "error reading system metrics: "+err.Error(), http.StatusInternalServerError)
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
