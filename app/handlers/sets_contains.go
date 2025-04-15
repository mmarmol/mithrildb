package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// setContainsHandler handles GET /documents/sets/contains
//
// @Summary      Check if element exists in set
// @Description  Checks whether a given element exists within a set-type document.
// @Tags         sets
// @Accept       json
// @Produce      json
// @Param        key     query     string  true   "Document key"
// @Param        cf      query     string  false  "Column family (default: 'default')"
// @Param        element query     string  true   "Element to check"
// @Success      200     {object}  map[string]interface{}  "Status and whether the element exists"
// @Failure      400     {object}  handlers.ErrorResponse  "Missing or invalid parameters"
// @Failure      404     {object}  handlers.ErrorResponse  "Document not found"
// @Failure      500     {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/sets/contains [get]
func setContainsHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf, err := getCfQueryParam(r)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		elementStr, err := getQueryParam(r, "element")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		opts := database.DefaultReadOptions
		if db.HasReadOptions(r) {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		contains, err := database.SetContains(cf, key, elementStr, opts)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"contains": contains,
		})
	}
}
