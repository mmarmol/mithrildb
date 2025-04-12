package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"mithrildb/db"
	"net/http"
	"time"
)

// Para el endpoint /stats, definimos una estructura con algunos datos
type Stats struct {
	Uptime string `json:"uptime"`
	DBPath string `json:"db_path"`
	// Podrías agregar más campos, como número de operaciones, tamaño de la base, etc.
}

// getQueryParam obtiene un parámetro de la URL y valida que no esté vacío.
func getQueryParam(r *http.Request, key string) (string, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return "", fmt.Errorf("missing '%s' parameter", key)
	}
	return val, nil
}

// GetHandler devuelve un handler HTTP para recuperar un valor desde la base de datos.
func GetHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("GET /get?key=%s", key)

		val, err := database.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if val == "" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(val))
	}
}

// PutHandler devuelve un handler HTTP para insertar o actualizar un valor.
func PutHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		val, err := getQueryParam(r, "val")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("PUT /put?key=%s", key)

		if err := database.PutDirect(key, val); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// DeleteHandler devuelve un handler HTTP para eliminar un valor.
func DeleteHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("DELETE /delete?key=%s", key)

		if err := database.DeleteDirect(key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// PingHandler devuelve un handler HTTP para verificar si el servicio está vivo.
func PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("pong"))
	}
}

// HealthHandler devuelve un estado de salud simple del servicio.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Podrías agregar chequeos adicionales, por ejemplo, revisar si la DB responde.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]string{"status": "healthy"}
		json.NewEncoder(w).Encode(resp)
	}
}

// StatsHandler devuelve estadísticas básicas del servidor.
func StatsHandler(dbPath string, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(startTime).Truncate(time.Second).String()
		stats := Stats{
			Uptime: uptime,
			DBPath: dbPath,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
