// handlers/create_family.go
package handlers

import (
	"encoding/json"
	"mithrildb/db"
	"net/http"
	"strings"
)

type createFamilyRequest struct {
	Name string `json:"name"`
}

// createFamilyHandler handles POST /families
func createFamilyHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createFamilyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		name := strings.TrimSpace(req.Name)
		if name == "" {
			respondWithError(w, http.StatusBadRequest, "missing or empty column family name")
			return
		}

		err := database.CreateFamily(name)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "created",
			"name":   name,
		})
	}
}
