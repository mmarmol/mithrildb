package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// listShiftHandler removes and returns the first element from a list document.
//
// @Summary      Shift list (remove first element)
// @Description  Removes and returns the first element of a list. Returns an error if the list is empty.
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        key  query     string  true  "Key of the list document"
// @Param        expiration  query  int  false  "Optional expiration. TTL in seconds (<= 30d) or absolute Unix timestamp (> 30d). Omit to keep existing expiration."
// @Param        cf   query     string  false "Column family (default: 'default')"
// @Param        sync        query bool false "Write option: wait for sync"
// @Param        disable_wal query bool false "Write option: disable WAL"
// @Param        no_slowdown query bool false "Write option: disable slowdown retries"
// @Success      200  {object}  map[string]interface{}  "Removed element from the list"
// @Failure      400  {object}  handlers.ErrorResponse  "Invalid input parameters"
// @Failure      404  {object}  handlers.ErrorResponse  "Document not found or list is empty"
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/lists/shift [post]
func listShiftHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
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

		result, err := database.ListShift(db.ListOpOptions{
			ColumnFamily: cf,
			Key:          key,
			WriteOptions: opts,
			Expiration:   expiration,
		})
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"element": result,
		})
	}
}
