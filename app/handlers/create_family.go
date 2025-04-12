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
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(req.Name)
		if name == "" {
			http.Error(w, "missing or empty column family name", http.StatusBadRequest)
			return
		}

		err := database.CreateFamily(name)
		if err != nil {
			if err == db.ErrFamilyExists {
				http.Error(w, fmt.Sprintf("column family '%s' already exists", name), http.StatusConflict)
				return
			}
			http.Error(w, fmt.Sprintf("failed to create column family: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "created",
			"name":   name,
		})
	}
}
