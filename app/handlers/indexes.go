package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"mithrildb/db"
	"mithrildb/indexes"
	"mithrildb/model"

	"github.com/gorilla/mux"
)

// createIndexHandler creates a new secondary index.
//
// @Summary      Create index
// @Description  Defines a new secondary index with projection and condition.
// @Tags         indexes
// @Accept       json
// @Produce      json
// @Param        index  body   model.IndexDefinition  true  "Index definition"
// @Success      201  {object}  map[string]string  "Index created successfully"
// @Failure      400  {object}  handlers.ErrorResponse  "Invalid index definition"
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /indexes [post]
func handleCreateIndex(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var def model.IndexDefinition
		if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if err := indexes.SaveIndex(database, &def); err != nil {
			respondWithError(w, http.StatusBadRequest, "failed to save index: "+err.Error())
			return
		}

		respondWithJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
	}
}

// listIndexesHandler returns all defined secondary indexes.
//
// @Summary      List indexes
// @Description  Retrieves all existing secondary index definitions.
// @Tags         indexes
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "List of index definitions"
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /indexes [get]
func handleListIndexes(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexes, err := indexes.ListIndexes(database)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to list indexes: "+err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, indexes)
	}
}

// getIndexHandler returns a single index definition by name.
//
// @Summary      Get index
// @Description  Retrieves the definition of a specific secondary index.
// @Tags         indexes
// @Accept       json
// @Produce      json
// @Param        name  path   string  true  "Index name"
// @Success      200  {object}  model.IndexDefinition  "Index definition"
// @Failure      400  {object}  handlers.ErrorResponse  "Invalid index name"
// @Failure      404  {object}  handlers.ErrorResponse  "Index not found"
// @Router       /indexes/{name} [get]
func handleGetIndex(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		index, err := indexes.GetIndex(database, name)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				respondWithError(w, http.StatusNotFound, "index not found")
			} else {
				respondWithError(w, http.StatusBadRequest, "error: "+err.Error())
			}
			return
		}

		respondWithJSON(w, http.StatusOK, index)
	}
}

// deleteIndexHandler removes a secondary index by name.
//
// @Summary      Delete index
// @Description  Deletes a secondary index and its definition from the system.
// @Tags         indexes
// @Accept       json
// @Produce      json
// @Param        name  path   string  true  "Index name"
// @Success      200  {object}  map[string]string  "Index deleted successfully"
// @Failure      400  {object}  handlers.ErrorResponse  "Invalid index name"
// @Failure      500  {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /indexes/{name} [delete]
func handleDeleteIndex(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		if err := indexes.DeleteIndex(database, name); err != nil {
			respondWithError(w, http.StatusBadRequest, "failed to delete index: "+err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
