package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"strconv"
)

// listKeysHandler handles GET /keys
//
// @Summary      List document keys
// @Description  Returns a list of keys within the specified column family, optionally filtered by prefix and pagination options.
// @Tags         keys
// @Produce      json
// @Param        cf           query  string  false  "Column family (default: 'default')"
// @Param        prefix       query  string  false  "Only return keys with this prefix"
// @Param        start_after  query  string  false  "Return keys after this key (for pagination)"
// @Param        limit        query  int     false  "Maximum number of keys to return (default: 100)"
// @Param        fill_cache   query  bool    false  "Whether to fill RocksDB read cache"
// @Param        read_tier    query  string  false  "RocksDB read tier (e.g. 'all', 'cache-only')"
// @Success      200  {array}  string
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/keys [get]
func listKeysHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := getCfQueryParam(r)

		prefix := r.URL.Query().Get("prefix")
		startAfter := r.URL.Query().Get("start_after")
		limitStr := r.URL.Query().Get("limit")

		limit := 100
		if limitStr != "" {
			if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
				limit = val
			}
		}

		// Read options
		opts := database.DefaultReadOptions
		override := db.HasReadOptions(r)
		if override {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		keys, err := database.ListKeys(cf, prefix, startAfter, limit, opts)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(keys)
	}
}
