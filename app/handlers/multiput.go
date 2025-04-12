package handlers

import (
	"encoding/json"
	"fmt"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/model"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func MultiPutHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode body: map of key -> { value, type }
		var payload map[string]struct {
			Value interface{} `json:"value"`
			Type  string      `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if len(payload) == 0 {
			http.Error(w, "empty payload", http.StatusBadRequest)
			return
		}

		cf := r.URL.Query().Get("cf")
		if cf == "" {
			cf = "default"
		}

		// Optional global expiration
		expiration := int64(0)
		if ttlStr := r.URL.Query().Get("expiration"); ttlStr != "" {
			if ttlVal, err := strconv.ParseInt(ttlStr, 10, 64); err == nil && ttlVal >= 0 {
				expiration = ttlVal
			}
		}

		// Write options
		opts := database.DefaultWriteOptions
		override := r.URL.Query().Has("sync") || r.URL.Query().Has("disable_wal") || r.URL.Query().Has("no_slowdown")
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		now := time.Now()
		docs := make(map[string]*model.Document)
		batch := make(map[string]interface{})

		for key, entry := range payload {
			if entry.Type == "" {
				entry.Type = model.DocTypeJSON
			}
			if err := model.ValidateValue(entry.Value, entry.Type); err != nil {
				http.Error(w, fmt.Sprintf("invalid value for key '%s': %v", key, err), http.StatusBadRequest)
				return
			}

			doc := &model.Document{
				Key:   key,
				Value: entry.Value,
				Meta: model.Metadata{
					Rev:        uuid.NewString(),
					Type:       entry.Type,
					Expiration: expiration,
					UpdatedAt:  now,
				},
			}
			docs[key] = doc
			batch[key] = doc
		}

		if err := database.MultiPut(cf, batch, opts); err != nil {
			http.Error(w, "multi put failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(docs)
	}
}
