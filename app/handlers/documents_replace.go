package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// documentReplaceHandler replaces an existing document by key, optionally using CAS.
//
// @Summary      Replace an existing document
// @Description  Replaces a document if it already exists. Fails if the key does not exist. Supports CAS for concurrency control.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        key   query     string                 true  "Document key"
// @Param        cf    query     string                 false "Column family (default: 'default')"
// @Param        expiration  query  int  false  "Optional expiration. TTL in seconds (<= 30d) or absolute Unix timestamp (> 30d). Omit to keep existing expiration."
// @Param        type  query     string                 false "Document type (e.g. 'json', 'counter', 'list')"
// @Param        cas   query     string                 false "CAS (revision) for concurrency control"
// @Param        body  body      map[string]interface{} true  "New value for the document"
// @Success      200   {object}  model.Document
// @Failure      400   {object}  handlers.ErrorResponse "Invalid request or missing value"
// @Failure      404   {object}  handlers.ErrorResponse "Key not found or column family missing"
// @Failure      409   {object}  handlers.ErrorResponse "CAS mismatch"
// @Failure      500   {object}  handlers.ErrorResponse "Internal server error"
// @Router       /documents/replace [post]
func documentReplaceHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Required: key
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		cf, err := getCfQueryParam(r)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		expiration, err := getExpirationQueryParam(r)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		docType := getDocTypeQueryParam(r)
		cas := getCasQueryParam(r)

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

		// Attempt replace
		doc, err := database.ReplaceDocument(db.DocumentWriteOptions{
			ColumnFamily: cf,
			Key:          key,
			Value:        body.Value,
			Cas:          cas,
			Type:         docType,
			Expiration:   expiration,
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
