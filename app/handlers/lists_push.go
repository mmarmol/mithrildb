package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// listElementRequest represents a request body to push or unshift an element into a list.
//
// Example:
//
//	{"element": "myValue"}
type listElementRequest struct {
	// Element to add to the list (can be string, number, object, etc.)
	Element interface{} `json:"element"`
}

// listPushHandler appends an element to the end of a list document.
//
// @Summary      Push element to list
// @Description  Adds an element to the end of an existing list-type document.
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        key  query     string              true  "Key of the list document"
// @Param        cf   query     string              false "Column family (default: 'default')"
// @Param        expiration  query  int  false  "Optional expiration. TTL in seconds (<= 30d) or absolute Unix timestamp (> 30d). Omit to keep existing expiration."
// @Param        body body      listElementRequest  true  "Element to add"
// @Param        sync          query  boolean false "Write option: sync write to disk"
// @Param        disable_wal   query  boolean false "Write option: disable write-ahead log"
// @Param        no_slowdown   query  boolean false "Write option: disable slowdown on write buffer full"
// @Success      200  {object}  map[string]string   "Status message"
// @Failure      400  {object}  handlers.ErrorResponse  "Invalid request or JSON body"
// @Failure      404  {object}  handlers.ErrorResponse  "Document not found"
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/lists/push [post]
func listPushHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
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

		var req listElementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		_, err = database.PushToList(db.ListPushOptions{
			ListOpOptions: db.ListOpOptions{
				ColumnFamily: cf,
				Key:          key,
				WriteOptions: opts,
				Expiration:   expiration,
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
