package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"mithrildb/db"
	"mithrildb/model"
	"net/http"
	"strconv"
	"time"
)

// ErrorResponse represents a standardized error message returned by the API.
// It includes a human-readable explanation of the error.
//
// Example:
//
//	{
//	  "error": "key not found"
//	}
type ErrorResponse struct {
	Error string `json:"error"`
}

func respondWithNotAllowed(w http.ResponseWriter) {
	respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
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

func getCfQueryParam(r *http.Request) (string, error) {
	cf := r.URL.Query().Get("cf")
	if cf == "" {
		cf = "default"
	}
	if !db.IsValidUserCF(cf) {
		return "", db.ErrInvalidUserColumnFamily
	}
	return cf, nil
}

func getDocTypeQueryParam(r *http.Request) string {
	docType := r.URL.Query().Get("type")
	if docType == "" {
		docType = model.DocTypeJSON
	}
	return docType
}

func getExpirationQueryParam(r *http.Request) (*int64, error) {
	param := r.URL.Query().Get("expiration")
	return parseExpirationParam(param)
}

// ParseExpirationParam interprets TTL or timestamp based on Couchbase logic
func parseExpirationParam(ttlStr string) (*int64, error) {
	if ttlStr == "" {
		return nil, nil // no expiration param sent
	}
	ttl, err := strconv.ParseInt(ttlStr, 10, 64)
	if err != nil {
		return nil, model.ErrInvalidExpiration
	}

	const thirtyDays = 60 * 60 * 24 * 30

	if ttl < 1 {
		zero := int64(0)
		return &zero, nil
	}

	var exp int64
	if ttl <= thirtyDays {
		exp = time.Now().Unix() + ttl
	} else {
		exp = ttl
	}
	return &exp, nil
}

func mapErrorToResponse(err error) (int, string) {
	switch {
	case errors.Is(err, db.ErrInvalidColumnFamily):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, model.ErrInvalidDocumentKey):
		return http.StatusBadRequest, err.Error()
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
	case errors.Is(err, db.ErrInvalidUserColumnFamily):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func mapAndRespondWithError(w http.ResponseWriter, err error) {
	status, msg := mapErrorToResponse(err)
	respondWithError(w, status, msg)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
