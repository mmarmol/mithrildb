package handlers

import (
	"encoding/json"
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
			respondWithErrInvalidJSONBody(w)
			return
		}
		if len(payload) == 0 {
			respondWithError(w, http.StatusBadRequest, "empty payload")
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
			if err == db.ErrInvalidColumnFamily {
				respondWithErrInvalidColumnFamily(w, cf)
				return
			}
			respondWithError(w, http.StatusInternalServerError, "multi put failed: "+err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(docs)
	}
}
