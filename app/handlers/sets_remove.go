package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// setRemoveHandler handles POST /documents/sets/remove
//
// @Summary      Remove element from set
// @Description  Removes a specific element from a set-type document.
// @Tags         sets
// @Accept       json
// @Produce      json
// @Param        key     query     string  true   "Document key"
// @Param        cf      query     string  false  "Column family (default: 'default')"
// @Param        body    body      map[string]interface{}  true  "Element to remove from the set"
// @Success      200     {object}  map[string]string        "Success message"
// @Failure      400     {object}  handlers.ErrorResponse   "Invalid request or missing parameters"
// @Failure      404     {object}  handlers.ErrorResponse   "Document not found"
// @Failure      500     {object}  handlers.ErrorResponse   "Internal server error"
// @Router       /documents/sets/remove [post]
func setRemoveHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := getCfQueryParam(r)
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
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

		_, err = database.SetRemove(db.SetOpOptions{
			ColumnFamily: cf,
			Key:          key,
			WriteOptions: opts,
		},
			req.Element)
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
