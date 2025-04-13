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

		cf := getCfQueryParam(r)

		// Write options
		opts := database.DefaultWriteOptions
		override := db.HasWriteOptions(r)
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
					Rev:       uuid.NewString(),
					Type:      entry.Type,
					UpdatedAt: now,
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
