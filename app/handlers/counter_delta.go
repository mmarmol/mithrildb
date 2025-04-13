package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

type incrementRequest struct {
	Delta int64 `json:"delta"`
}

type incrementResponse struct {
	Old int64 `json:"old"`
	New int64 `json:"new"`
}

// CounterIncrementHandler handles POST /counters/delta
func DeltaCountertHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get required 'key' query param
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		cf := getCfQueryParam(r)

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

		oldVal, newVal, err := database.DeltaCounter(cf, key, req.Delta, opts)
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
