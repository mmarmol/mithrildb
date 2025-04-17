package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// incrementRequest represents the JSON body for a counter modification request.
//
// It expects a non-zero numeric value in the `delta` field. It can be positive or negative.
//
// Examples:
//
//	{"delta": 5}     // Increment by 5
//	{"delta": -2}    // Decrement by 2
type incrementRequest struct {
	// The amount to increment (positive) or decrement (negative) the counter.
	Delta int64 `json:"delta"`
}

// incrementResponse represents the response for a counter delta operation.
type incrementResponse struct {
	Old int64 `json:"old"` // Previous value
	New int64 `json:"new"` // New value
}

// deltaCounterHandler applies a delta operation to a counter document.
//
// @Summary      Modify counter
// @Description  Increments or decrements a counter document by a given integer value.
// @Tags         counters
// @Accept       json
// @Produce      json
// @Param        key   query     string               true  "Document key"
// @Param        cf    query     string               false "Column family (default: 'default')"
// @Param        expiration  query  int  false  "Expiration time in seconds (TTL <= 30d or Unix timestamp)"
// @Param        body  body      incrementRequest     true  "Delta value for increment or decrement"
// @Success      200   {object}  incrementResponse
// @Failure      400   {object}  handlers.ErrorResponse  "Invalid parameters or JSON body"
// @Failure      404   {object}  handlers.ErrorResponse  "Document not found"
// @Failure      500   {object}  handlers.ErrorResponse  "Internal server error"
// @Router       /documents/counters/delta [post]
func deltaCountertHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get required 'key' query param
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		cf, err := getCfQueryParam(r)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		expiration, err := getExpirationQueryParam(r)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		// Parse JSON body with delta
		var req incrementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithErrInvalidJSONBody(w)
			return
		}
		if req.Delta == 0 {
			respondWithError(w, http.StatusBadRequest, "'delta' must be a non-zero integer")
			return
		}

		// Write options
		opts := database.DefaultWriteOptions
		override := db.HasWriteOptions(r)
		if override {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		oldVal, newVal, err := database.DeltaCounter(cf, key, req.Delta, *expiration, opts)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(incrementResponse{
			Old: oldVal,
			New: newVal,
		})
	}
}
