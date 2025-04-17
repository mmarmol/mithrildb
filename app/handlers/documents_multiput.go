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

// BulkPutRequestEntry represents a single entry in a bulk document insert request.
//
// Each entry includes a value and an optional document type.
// If the type is not provided, "json" will be assumed.
//
// Example:
//
//	{
//	  "user:1": { "value": { "name": "Alice" }, "type": "json" },
//	  "counter:1": { "value": 42, "type": "counter" }
//	}
type BulkPutRequestEntry struct {
	// The value to store for the document.
	// Can be a string, number, object, array, etc.
	Value interface{} `json:"value"`

	// Optional document type (e.g., "json", "counter", "list", "set").
	// If omitted, "json" is assumed.
	Type string `json:"type,omitempty"`
}

// bulkPutHandler stores multiple documents in a single request.
//
// @Summary      Bulk insert documents
// @Description  Stores multiple documents in a single call. Each document is defined by a key and a value/type pair.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        cf    query     string                              false  "Column family (defaults to 'default')"
// @Param        expiration  query  int  false  "Optional expiration. TTL in seconds (<= 30d) or absolute Unix timestamp (> 30d). Omit to keep existing expiration."
// @Param        body  body      map[string]handlers.BulkPutRequestEntry  true  "Map of key to value/type entry"
// @Success      200   {object}  map[string]model.Document
// @Failure      400   {object}  handlers.ErrorResponse  "Invalid input or empty payload"
// @Failure      500   {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/bulk/put [post]
func bulkPutHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]BulkPutRequestEntry
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondWithErrInvalidJSONBody(w)
			return
		}
		if len(payload) == 0 {
			respondWithError(w, http.StatusBadRequest, "empty payload")
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

		if err := database.MultiPut(cf, batch, expiration, opts); err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(docs)
	}
}
