package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// documentPutHandler stores a document with optional CAS and metadata.
//
// @Summary      Store or update a document
// @Description  Stores a document in the specified column family. Supports optimistic concurrency via CAS.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        key   query     string            true  "Document key"
// @Param        cf    query     string            false "Column family (default: 'default')"
// @Param        type  query     string            false "Document type (e.g. 'json', 'counter', 'list')"
// @Param        cas   query     string            false "CAS (revision) for concurrency control"
// @Param        body  body      map[string]interface{}  true  "Document value (JSON-encoded)"
// @Success      200   {object}  model.Document
// @Failure      400   {object}  handlers.ErrorResponse "Invalid input or missing value"
// @Failure      404   {object}  handlers.ErrorResponse "Column family not found"
// @Failure      409   {object}  handlers.ErrorResponse "CAS mismatch"
// @Failure      500   {object}  handlers.ErrorResponse "Internal server error"
// @Router       /documents [post]
func documentPutHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
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
