package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
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

		cf := getCfQueryParam(r)
		docType := getDocTypeQueryParam(r)

		// Read JSON body
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

		// Attempt insert
		doc, err := database.Insert(db.PutOptions{
			ColumnFamily: cf,
			Key:          key,
			Value:        body.Value,
			Type:         docType,
			WriteOptions: opts,
		})

		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		// Success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	}
}
