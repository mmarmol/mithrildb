package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/model"
	"net/http"
)

// PutHandler stores a document using the new document model with metadata and optional CAS.
func PutHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Required: key in query
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
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
			respondWithErrInvalidJSONBody(w)
			return
		}
		if body.Value == nil {
			respondWithErrMissingValue(w)
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
			if err == db.ErrInvalidColumnFamily {
				respondWithErrInvalidColumnFamily(w, cf)
				return
			}
			if err == db.ErrRevisionMismatch {
				respondWithError(w, http.StatusPreconditionFailed, err.Error())
				return
			}
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Respond with document
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	}
}
