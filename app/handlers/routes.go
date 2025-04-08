package handlers

import (
	"fmt"
	"mithrildb/config"
	"net/http"
)

func SetupRoutes() {
	cfg := config.LoadConfig() // Load configuration

	http.HandleFunc("/get", handleGet)
	http.HandleFunc("/put", handlePut)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/ping", handlePing)

	http.ListenAndServe(":"+fmt.Sprintf("%d", cfg.Port), nil)
}
