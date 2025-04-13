// handlers/create_family.go
package handlers

import (
	"encoding/json"
	"fmt"
	"mithrildb/db"
	"net/http"
	"strings"
)

type createFamilyRequest struct {
	Name string `json:"name"`
}

// CreateFamilyHandler handles POST /families
func CreateFamilyHandler(database *db.DB) http.HandlerFunc {
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
			if err == db.ErrFamilyExists {
				respondWithError(w, http.StatusConflict, fmt.Sprintf("column family '%s' already exists", name))
				return
			}
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create column family: %v", err))
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "created",
			"name":   name,
		})
	}
}
