// handlers/create_family.go
package handlers

import (
	"encoding/json"
	"mithrildb/db"
	"net/http"
	"strings"
)

// createFamilyRequest represents the body for creating a new column family.
//
// It requires a non-empty name for the new column family.
//
// Example:
//
//	{
//	  "name": "logs"
//	}
type createFamilyRequest struct {
	Name string `json:"name" example:"logs"`
}

// createFamilyHandler creates a new column family in the database.
//
// @Summary      Create column family
// @Description  Creates a new column family with the specified name.
// @Tags         families
// @Accept       json
// @Produce      json
// @Param        body  body      createFamilyRequest  true  "Name of the column family to create"
// @Success      201   {object}  map[string]string     "Created column family name"
// @Failure      400   {object}  handlers.ErrorResponse "Invalid or missing column family name"
// @Failure      409   {object}  handlers.ErrorResponse "Column family already exists"
// @Failure      500   {object}  handlers.ErrorResponse "Internal server error"
// @Router       /families [post]
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

		if !db.IsValidUserCF(name) {
			mapAndRespondWithError(w, db.ErrInvalidUserColumnFamily)
			return
		}

		err := database.CreateColumnFamily(name)
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
