package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/model"
	"net/http"
)

// InsertHandler handles POST /insert?key=...&cf=...&type=...
// The document is only inserted if it doesn't already exist.
func InsertHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Required: key
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Optional: column family
		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default"
		}

		// Optional: type
		docType := r.URL.Query().Get("type")
		if docType == "" {
			docType = model.DocTypeJSON
		}

		// Optional: expiration
		expiration := int64(0)

		// Read JSON body
		var body struct {
			Value interface{} `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.Value == nil {
			http.Error(w, "missing 'value' in body", http.StatusBadRequest)
			return
		}

		// Write options
		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		// Attempt insert
		doc, err := database.Insert(db.PutOptions{
			ColumnFamily: cf,
			Key:          key,
			Value:        body.Value,
			Type:         docType,
			Expiration:   expiration,
			WriteOptions: opts,
		})
		if err != nil {
			if err == db.ErrKeyAlreadyExists {
				http.Error(w, "key already exists", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	}
}
