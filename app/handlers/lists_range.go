package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"strconv"
)

// listRangeHandler retrieves a range of elements from a list document.
//
// @Summary      Get elements from a list
// @Description  Returns a slice of elements from a list document, based on start and end indices.
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        key   query     string  true  "Key of the list document"
// @Param        cf    query     string  false "Column family (default: 'default')"
// @Param        start query     int     true  "Start index (inclusive, 0-based)"
// @Param        end   query     int     true  "End index (inclusive, -1 for end of list)"
// @Param        fill_cache query bool false "Read option: whether to fill RocksDB cache"
// @Param        read_tier  query string false "Read option: 'all' or 'cache-only'"
// @Success      200  {object}  map[string]interface{}  "List content"
// @Failure      400  {object}  handlers.ErrorResponse  "Invalid input parameters"
// @Failure      404  {object}  handlers.ErrorResponse  "Document not found"
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/lists/range [get]
func listRangeHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		cf, err := getCfQueryParam(r)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		start, _ := strconv.Atoi(r.URL.Query().Get("start"))
		end, _ := strconv.Atoi(r.URL.Query().Get("end"))

		opts := database.DefaultReadOptions
		if db.HasReadOptions(r) {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		res, err := database.GetListRange(db.ListRangeOptions{
			ColumnFamily: cf,
			Key:          key,
			Start:        start,
			End:          end,
			ReadOptions:  opts,
		})
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"list": res,
		})
	}
}
