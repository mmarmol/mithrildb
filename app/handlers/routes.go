package handlers

import (
	"mithrildb/db"
	"net/http"
	"time"
)

// SetupRoutes registra todas las rutas HTTP con sus respectivos handlers.
func SetupRoutes(database *db.DB, dbPath string, startTime time.Time) {
	http.Handle("/get", GetHandler(database))
	http.Handle("/put", PutHandler(database))
	http.Handle("/delete", DeleteHandler(database))
	http.Handle("/ping", PingHandler())
	http.Handle("/health", HealthHandler())
	http.Handle("/stats", StatsHandler(dbPath, startTime))
}
