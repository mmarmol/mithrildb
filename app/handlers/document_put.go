package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
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

		cf := getCfQueryParam(r)
		docType := getDocTypeQueryParam(r)
		cas := getCasQueryParam(r)

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
		override := db.HasWriteOptions(r)
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
			Type:         docType,
			WriteOptions: opts,
		}

		// Execute put
		doc, err := database.PutWithOptions(putOpts)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		// Respond with document
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	}
}
