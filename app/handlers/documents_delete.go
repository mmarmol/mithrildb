package handlers

import (
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// documentDeleteHandler handles DELETE /documents
//
// @Summary      Delete a document
// @Description  Deletes a document by key within a specified column family.
// @Tags         documents
// @Produce      json
// @Param        key  query  string  true   "Document key to delete"
// @Param        cf   query  string  false  "Column family (default: 'default')"
// @Success      200  "Document successfully deleted"
// @Failure      400  {object}  handlers.ErrorResponse  "Missing or invalid parameters"
// @Failure      404  {object}  handlers.ErrorResponse  "Document or column family not found"
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents [delete]
func documentDeleteHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
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

		// Determine write options
		opts := database.DefaultWriteOptions
		override := db.HasWriteOptions(r)
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		// Call Delete with the specified column family
		if err := database.DeleteDirect(cf, key, opts); err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
