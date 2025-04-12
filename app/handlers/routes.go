package handlers

import (
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"time"
)

// SetupRoutes registra todas las rutas HTTP con sus respectivos handlers.
func SetupRoutes(database *db.DB, cfg config.AppConfig, startTime time.Time) {
	http.Handle("/get", GetHandler(database, cfg.ReadDefaults))
	http.Handle("/multiget", MultiGetHandler(database, cfg.ReadDefaults))
	http.Handle("/put", PutHandler(database, cfg.WriteDefaults))
	http.Handle("/multiput", MultiPutHandler(database, cfg.WriteDefaults))
	http.Handle("/delete", DeleteHandler(database, cfg.WriteDefaults))
	http.Handle("/ping", PingHandler())
	http.Handle("/list", ListHandler(database))
	http.Handle("/metrics", MetricsHandler(database, cfg, startTime))
	http.HandleFunc("/families", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ListFamiliesHandler(database)(w, r)
		case http.MethodPost:
			CreateFamilyHandler(database)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
