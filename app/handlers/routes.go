package handlers

import (
	"net/http"
)

func SetupRoutes() {
	http.HandleFunc("/get", handleGet)
	http.HandleFunc("/put", handlePut)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/ping", handlePing)
}
