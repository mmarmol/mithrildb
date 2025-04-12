package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/model"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// MultiPutHandler handles POST /multiput
// It accepts a map of key-value pairs and stores them as documents with metadata.
func MultiPutHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if len(payload) == 0 {
			http.Error(w, "empty key-value map", http.StatusBadRequest)
			return
		}

		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default"
		}

		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		docs := make(map[string]*model.Document, len(payload))
		now := time.Now()
		for k, v := range payload {
			doc := &model.Document{
				Key:   k,
				Value: v,
				Meta: model.Metadata{
					Rev:        uuid.NewString(),
					Type:       model.DocTypeJSON,
					UpdatedAt:  now,
					Expiration: 0,
				},
			}
			docs[k] = doc
		}

		// Create batch from prepared documents
		batch := make(map[string]string)
		for k, doc := range docs {
			data, err := json.Marshal(doc)
			if err != nil {
				http.Error(w, "failed to encode document: "+err.Error(), http.StatusInternalServerError)
				return
			}
			batch[k] = string(data)
		}

		if err := database.MultiPut(cf, batch, opts); err != nil {
			http.Error(w, "multi put failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(docs)
	}
}
