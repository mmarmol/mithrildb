package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"mithrildb/db"
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

func mapErrorToResponse(err error) (int, string) {
	switch {
	case errors.Is(err, db.ErrInvalidColumnFamily):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, db.ErrKeyNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, db.ErrRevisionMismatch):
		return http.StatusPreconditionFailed, err.Error()
	case errors.Is(err, db.ErrKeyAlreadyExists):
		return http.StatusConflict, err.Error()
	case errors.Is(err, db.ErrInvalidListType):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, db.ErrEmptyList):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, model.ErrInvalidCounterValue):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, model.ErrInvalidCounterType):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, db.ErrInvalidSetType):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, db.ErrFamilyExists):
		return http.StatusConflict, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func mapAndRespondWithError(w http.ResponseWriter, err error) {
	status, msg := mapErrorToResponse(err)
	respondWithError(w, status, msg)
}
