package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// documentGetHandler handles GET /documents
//
// @Summary      Retrieve a document
// @Description  Retrieves a document by key, including its value and metadata.
// @Tags         documents
// @Produce      json
// @Param        key  query     string  true   "Document key"
// @Param        cf   query     string  false  "Column family (default: 'default')"
// @Param        fill_cache query bool false "Optional RocksDB fill cache read option"
// @Param        read_tier query string false "Optional RocksDB read tier (e.g. 'all', 'cache-only')"
// @Success      200  {object}  model.Document
// @Failure      400  {object}  handlers.ErrorResponse  "Missing or invalid key"
// @Failure      404  {object}  handlers.ErrorResponse  "Document not found"
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents [get]
func documentGetHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
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

		opts := database.DefaultReadOptions
		override := db.HasReadOptions(r)
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		doc, err := database.GetDocument(db.DocumentReadOptions{
			ColumnFamily: cf,
			Key:          key,
			ReadOptions:  opts,
		})
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	}
}
