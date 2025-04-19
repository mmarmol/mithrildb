package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// @Summary Insert a document
// @Description Insert a new document only if the key does not already exist
// @Tags documents
// @Accept json
// @Produce json
// @Param key query string true "Document key"
// @Param cf query string false "Column family (defaults to 'default')"
// @Param type query string false "Document type (json, counter, list, set)"
// @Param        expiration  query  int  false  "Optional expiration. TTL in seconds (<= 30d) or absolute Unix timestamp (> 30d). Omit to store without expiration."
// @Param sync query bool false "Write option: sync"
// @Param disable_wal query bool false "Write option: disable WAL"
// @Param no_slowdown query bool false "Write option: no slowdown"
// @Param document body map[string]interface{} true "Document value body. Must contain 'value' field"
// @Success 200 {object} model.Document
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 409 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /documents/insert [post]
func documentInsertHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
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
		doc, err := database.InsertDocument(db.DocumentWriteOptions{
			ColumnFamily: cf,
			Key:          key,
			Value:        body.Value,
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
