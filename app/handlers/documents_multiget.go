package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/model"
	"net/http"
)

// multiGetRequest defines the structure of the JSON body expected in a MultiGet request.
type multiGetRequest struct {
	Keys []string `json:"keys"`
}

// MultiGetResponse represents the result of a multiget.
// @Description Map of keys to documents or null for missing entries.
type MultiGetResponse map[string]*model.Document

// bulkGetHandler handles POST /documents/bulk/get
//
// @Summary      Bulk document fetch
// @Description  Retrieves multiple documents with metadata by key. Missing keys will be returned with null values.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        cf    query     string             false  "Column family (default: 'default')"
// @Param        body  body      multiGetRequest    true   "List of keys to retrieve"
// @Success      200   {object}  MultiGetResponse
// @Failure      400   {object}  ErrorResponse  "Invalid JSON or missing key list"
// @Failure      500   {object}  ErrorResponse  "Internal server error"
// @Router       /documents/bulk/get [post]
func bulkGetHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req multiGetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithErrInvalidJSONBody(w)
			return
		}
		if len(req.Keys) == 0 {
			respondWithError(w, http.StatusBadRequest, "empty key list")
			return
		}

		cf := getCfQueryParam(r)

		opts := database.DefaultReadOptions
		override := db.HasReadOptions(r)
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		// Perform the multi-get operation and return document objects (not just values)
		result, err := database.MultiGet(cf, req.Keys, opts)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
