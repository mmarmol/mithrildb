package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"mithrildb/rocks"
)

var (
	db    *rocks.DB
	cfg   *rocks.Config
	cfgMu sync.RWMutex
)

func main() {
	var err error

	cfg = rocks.DefaultConfig()
	db, err = rocks.Open("/data/db", cfg)
	if err != nil {
		log.Fatal("Error abriendo base de datos:", err)
	}
	defer db.Close()

	http.HandleFunc("/get", withMetrics(handleGet))
	http.HandleFunc("/put", withMetrics(handlePut))
	http.HandleFunc("/delete", withMetrics(handleDelete))
	http.HandleFunc("/config/runtime", handleRuntimeConfig)
	http.HandleFunc("/config/static", handleStaticConfig)
	http.HandleFunc("/metrics", handleMetrics)

	log.Println("Servidor escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// GET /get?key=foo
func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val, err := db.Get(key)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(val))
}

// PUT /put?key=foo&val=bar
func handlePut(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("val")
	if err := db.Put(key, val); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if err := db.Delete(key); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

func handleRuntimeConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cfgMu.Lock()
		defer cfgMu.Unlock()
		var newCfg rocks.RuntimeConfig
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			http.Error(w, "JSON inválido", 400)
			return
		}
		cfg.ApplyRuntimeConfig(newCfg)
		db.ApplyRuntimeConfig(newCfg)
		w.WriteHeader(200)
	} else if r.Method == http.MethodGet {
		cfgMu.RLock()
		defer cfgMu.RUnlock()
		json.NewEncoder(w).Encode(cfg.RuntimeConfig)
	}
}

func handleStaticConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		cfgMu.RLock()
		defer cfgMu.RUnlock()
		json.NewEncoder(w).Encode(cfg.StaticConfig)
	} else {
		http.Error(w, "Solo lectura", 405)
	}
}

var (
	metricsMu   sync.Mutex
	getCount    int
	putCount    int
	deleteCount int
)

func withMetrics(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsMu.Lock()
		switch r.URL.Path {
		case "/get":
			getCount++
		case "/put":
			putCount++
		case "/delete":
			deleteCount++
		}
		metricsMu.Unlock()
		h(w, r)
	}
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	fmt.Fprintf(w, "get_requests %d\nput_requests %d\ndelete_requests %d\n", getCount, putCount, deleteCount)
}
