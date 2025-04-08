package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"mithrildb/db"
	"mithrildb/rocks"
	"mithrildb/services"
)

var (
	cfgMu sync.RWMutex
)

// Consolidar manejador de configuración
func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cfgMu.Lock()
		defer cfgMu.Unlock()
		var newCfg rocks.Config
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			http.Error(w, "JSON inválido", 400)
			return
		}

		// Traducción de configuración aquí, antes de la aplicación
		db.Cfg = &newCfg
		if err := db.DB.ApplyConfig(newCfg); err != nil {
			http.Error(w, "Error al aplicar configuración", 500)
			return
		}

		w.WriteHeader(200)
	} else if r.Method == http.MethodGet {
		cfgMu.RLock()
		defer cfgMu.RUnlock()
		json.NewEncoder(w).Encode(db.Cfg)
	} else {
		http.Error(w, "Método no permitido", 405)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val, err := db.DB.Get(key)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(val))
}

func handlePut(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("val")
	if err := db.DB.Put(key, val); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if err := db.DB.Delete(key); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

var withMetrics = services.WithMetrics
var handleMetrics = services.HandleMetrics
