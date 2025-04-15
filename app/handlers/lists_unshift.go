package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// listUnshiftHandler adds an element to the beginning of a list document.
//
// @Summary      Unshift list (add to start)
// @Description  Adds a new element to the beginning of a list document. The element can be of any JSON-compatible type.
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        key   query     string                 true  "Key of the list document"
// @Param        cf    query     string                 false "Column family (default: 'default')"
// @Param        body  body      listElementRequest     true  "Element to insert at the beginning"
// @Param        sync        query bool false "Write option: wait for sync"
// @Param        disable_wal query bool false "Write option: disable WAL"
// @Param        no_slowdown query bool false "Write option: disable slowdown retries"
// @Success      200   {object}  map[string]interface{}  "Insertion successful"
// @Failure      400   {object}  handlers.ErrorResponse  "Invalid input or JSON body"
// @Failure      404   {object}  handlers.ErrorResponse  "Document not found"
// @Failure      500   {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/lists/unshift [post]
func listUnshiftHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
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

		var req struct {
			Element interface{} `json:"element"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Element == nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		_, err = database.ListUnshift(db.ListPushOptions{
			ListOpOptions: db.ListOpOptions{
				ColumnFamily: cf,
				Key:          key,
				WriteOptions: opts,
			},
			Element: req.Element,
		})
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}
}
