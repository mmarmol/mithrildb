package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// SetElementRequest represents a request to add or remove an element in a set.
// @Description Request body containing the element to operate with.
type SetElementRequest struct {
	Element interface{} `json:"element"`
}

// setAddHandler handles POST /documents/sets/add
//
// @Summary      Add element to set
// @Description  Adds a new element to a document of type "set". If the element already exists, it will not be duplicated.
// @Tags         sets
// @Accept       json
// @Produce      json
// @Param        key   query     string                  true  "Document key"
// @Param        cf    query     string                  false "Column family (default: 'default')"
// @Param        body  body      handlers.SetElementRequest true "Element to add to the set"
// @Success      200   {object}  map[string]string       "Operation successful"
// @Failure      400   {object}  handlers.ErrorResponse  "Invalid parameters or body"
// @Failure      404   {object}  handlers.ErrorResponse  "Document not found"
// @Failure      500   {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/sets/add [post]
func setAddHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := getCfQueryParam(r)
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var req SetElementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Element == nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		_, err = database.SetAdd(db.SetOpOptions{
			ColumnFamily: cf,
			Key:          key,
			WriteOptions: opts,
		},
			req.Element,
		)
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
