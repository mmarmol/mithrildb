package handlers

import (
	"net/http"

	"mithrildb/services"
)

func SetupRoutes() {
	http.HandleFunc("/get", services.WithMetrics(handleGet))
	http.HandleFunc("/put", services.WithMetrics(handlePut))
	http.HandleFunc("/delete", services.WithMetrics(handleDelete))
	http.HandleFunc("/config", handleConfig) // Ruta consolidada para configuración
	http.HandleFunc("/metrics", services.HandleMetrics)
}
