package handlers

import (
	"encoding/json"
	"fmt"
	"mithrildb/model"
	"net/http"
)

// ErrorResponse represents a standard JSON error message.
type ErrorResponse struct {
	Error string `json:"error"`
}

// respondWithError sends a JSON error response with the given status and message.
func respondWithError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func respondWithErrInvalidColumnFamily(w http.ResponseWriter, cf string) {
	respondWithError(w, http.StatusNotFound, fmt.Sprintf("column family '%s' does not exists", cf))
}

func respondWithErrInvalidJSONBody(w http.ResponseWriter) {
	respondWithError(w, http.StatusBadRequest, "invalid JSON body")
}

func respondWithErrMissingValue(w http.ResponseWriter) {
	respondWithError(w, http.StatusBadRequest, "missing or null 'value' in body")
}

func getQueryParam(r *http.Request, key string) (string, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return "", fmt.Errorf("missing '%s' parameter", key)
	}
	return val, nil
}

func getCasQueryParam(r *http.Request) string {
	return r.URL.Query().Get("cas")
}

func getCfQueryParam(r *http.Request) string {
	cf := r.URL.Query().Get("cf")
	if cf == "" {
		cf = "default"
	}
	return cf
}

func getDocTypeQueryParam(r *http.Request) string {
	docType := r.URL.Query().Get("type")
	if docType == "" {
		docType = model.DocTypeJSON
	}
	return docType
}
