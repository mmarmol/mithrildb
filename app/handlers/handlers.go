package handlers

import (
	"encoding/json"
	"fmt"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/model"
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

// GetHandler retrieves a full document including metadata.
func GetHandler(database *db.DB, defaults config.ReadOptionsConfig, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default"
		}

		opts := database.DefaultReadOptions
		override := r.URL.Query().Has("fill_cache") || r.URL.Query().Has("read_tier")
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		doc, err := database.Get(cf, key, opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if doc == nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	}
}

// PutHandler stores a document using the new document model with metadata and optional CAS.
func PutHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Required: key in query
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Optional: column family (defaults to "default")
		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default"
		}

		// Optional: CAS and type hint
		cas := r.URL.Query().Get("cas")
		typeHint := r.URL.Query().Get("type")
		if typeHint == "" {
			typeHint = model.DocTypeJSON
		}

		// Optional: expiration (not parsed yet, placeholder)
		expiration := int64(0)

		// Parse the body (value can be any valid JSON type)
		var body struct {
			Value interface{} `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.Value == nil {
			http.Error(w, "missing or null 'value' in body", http.StatusBadRequest)
			return
		}

		// Write options
		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		// Build put options
		putOpts := db.PutOptions{
			ColumnFamily: cf,
			Key:          key,
			Value:        body.Value,
			Cas:          cas,
			Type:         typeHint,
			Expiration:   expiration,
			WriteOptions: opts,
		}

		// Execute put
		doc, err := database.PutWithOptions(putOpts)
		if err != nil {
			if err == db.ErrRevisionMismatch {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusPreconditionFailed)
				json.NewEncoder(w).Encode(map[string]string{
					"error": err.Error(),
				})
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Respond with document
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	}
}

func DeleteHandler(database *db.DB, defaults config.WriteOptionsConfig, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default" // Default column family
		}

		// Determine write options
		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		// Call Delete with the specified column family
		if err := database.DeleteDirect(cf, key, opts); err != nil {
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
