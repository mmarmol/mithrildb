package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// documentTouchHandler updates only the expiration of an existing document.
//
// @Summary      Update document expiration
// @Description  Updates the expiration time of an existing document without modifying its content. Fails if the key does not exist.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        key         query  string true  "Document key"
// @Param        cf          query  string false "Column family (default: 'default')"
// @Param        expiration  query  int  false  "Optional expiration. TTL in seconds (<= 30d) or absolute Unix timestamp (> 30d). Omit to keep existing expiration."
// @Success      200 {object} model.Document
// @Failure      400 {object} handlers.ErrorResponse "Invalid request"
// @Failure      404 {object} handlers.ErrorResponse "Key not found or already expired"
// @Failure      500 {object} handlers.ErrorResponse "Internal server error"
// @Router       /documents/touch [post]
func documentTouchHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Required: key
		key, err := getQueryParam(r, "key")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Optional: CF and expiration
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

		// Write options
		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		// Perform touch
		doc, err := database.TouchDocument(db.DocumentWriteOptions{
			ColumnFamily: cf,
			Key:          key,
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
