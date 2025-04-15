// handlers/families.go
package handlers

import (
	"encoding/json"
	"mithrildb/db"
	"net/http"
)

// listFamiliesHandler returns a list of all column families in the database.
//
// @Summary      List column families
// @Description  Retrieves the names of all available column families.
// @Tags         families
// @Produce      json
// @Success      200  {array}   string
// @Failure      500  {object}  handlers.ErrorResponse "Internal server error"
// @Router       /families [get]
func listFamiliesHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usrFamilies, _ := database.ListFamilyNames()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(usrFamilies)
	}
}
