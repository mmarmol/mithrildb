package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"strconv"
)

// listDocumentsHandler handles GET /documents/list with optional prefix, pagination and read options.
//
// @Summary      List documents
// @Description  Returns a map of documents filtered by optional prefix and paginated using start_after and limit.
// @Tags         documents
// @Produce      json
// @Param        cf           query     string  false  "Column family (default: 'default')"
// @Param        prefix       query     string  false  "Filter documents whose keys begin with this prefix"
// @Param        start_after  query     string  false  "Skip documents until this key (exclusive)"
// @Param        limit        query     int     false  "Maximum number of documents to return (default: 100)"
// @Param        fill_cache   query     bool    false  "Whether to fill RocksDB read cache"
// @Param        read_tier    query     string  false  "RocksDB read tier (e.g., 'all', 'cache-only')"
// @Success      200  {object}  map[string]model.Document
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/list [get]
func listDocumentsHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf, err := getCfQueryParam(r)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		prefix := r.URL.Query().Get("prefix")
		startAfter := r.URL.Query().Get("start_after")
		limitStr := r.URL.Query().Get("limit")

		limit := 100
		if limitStr != "" {
			if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
				limit = val
			}
		}

		// Read options (use default for now)
		opts := database.DefaultReadOptions
		override := db.HasReadOptions(r)
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		// First list keys
		keys, err := database.ListDocumentKeys(db.KeyListOptions{
			ColumnFamily: cf,
			Prefix:       prefix,
			StartAfter:   startAfter,
			Limit:        limit,
			ReadOptions:  opts,
		})
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		// Then fetch documents for those keys
		docs, err := database.BulkGetDocuments(db.BulkReadOptions{
			ColumnFamily: cf,
			Keys:         keys,
			ReadOptions:  opts,
		})
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		// Return map of key => document
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(docs)
	}
}
