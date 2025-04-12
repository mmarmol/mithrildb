package handlers

import (
	"fmt"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

type Stats struct {
	Uptime string `json:"uptime"`
	DBPath string `json:"db_path"`
}

func getQueryParam(r *http.Request, key string) (string, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return "", fmt.Errorf("missing '%s' parameter", key)
	}
	return val, nil
}

func GetHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		opts := database.DefaultReadOptions
		override := r.URL.Query().Has("fill_cache") || r.URL.Query().Has("read_tier")
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		value, err := database.TransactionDB.Get(opts, []byte(key))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer value.Free()

		if value.Size() == 0 {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write(value.Data())
	}
}

func PutHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
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

		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		if err := database.TransactionDB.Put(opts, []byte(key), []byte(val)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func DeleteHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		if err := database.TransactionDB.Delete(opts, []byte(key)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("pong"))
	}
}
