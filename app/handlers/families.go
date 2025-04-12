// handlers/families.go
package handlers

import (
	"encoding/json"
	"mithrildb/db"
	"net/http"
)

func ListFamiliesHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := database.ListFamilyNames()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(names)
	}
}
