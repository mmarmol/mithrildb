package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// listPopHandler removes and returns the last element from a list document.
//
// @Summary      Pop element from list
// @Description  Removes and returns the last element of a list-type document.
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        key  query     string  true  "Key of the list document"
// @Param        cf   query     string  false "Column family (default: 'default')"
// @Param        expiration  query  int  false  "Optional expiration. TTL in seconds (<= 30d) or absolute Unix timestamp (> 30d). Omit to keep existing expiration."
// @Param        sync          query  boolean false "Write option: sync write to disk"
// @Param        disable_wal   query  boolean false "Write option: disable write-ahead log"
// @Param        no_slowdown   query  boolean false "Write option: disable slowdown on write buffer full"
// @Success      200  {object}  map[string]interface{}  "Returns the popped element"
// @Failure      400  {object}  handlers.ErrorResponse  "Invalid request or missing key"
// @Failure      404  {object}  handlers.ErrorResponse  "Document not found"
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/lists/pop [post]
func listPopHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
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
		expiration, err := getExpirationQueryParam(r)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		res, err := database.ListPop(db.ListOpOptions{
			ColumnFamily: cf,
			Key:          key,
			Expiration:   expiration,
			WriteOptions: opts,
		})
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"element": res,
		})
	}
}
