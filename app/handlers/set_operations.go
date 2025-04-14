package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
)

// SetAddHandler handles POST /documents/sets/add
func SetAddHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := getCfQueryParam(r)
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var req struct {
			Element interface{} `json:"element"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Element == nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		_, err = database.SetAdd(db.SetOpOptions{
			ColumnFamily: cf,
			Key:          key,
			WriteOptions: opts,
		},
			req.Element,
		)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}
}

// SetRemoveHandler handles POST /documents/sets/remove
func SetRemoveHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := getCfQueryParam(r)
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var req struct {
			Element interface{} `json:"element"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Element == nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		_, err = database.SetRemove(db.SetOpOptions{
			ColumnFamily: cf,
			Key:          key,
			WriteOptions: opts,
		},
			req.Element)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}
}

// SetContainsHandler handles GET /documents/sets/contains
func SetContainsHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cf := getCfQueryParam(r)
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		elementStr, err := getQueryParam(r, "element")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		opts := database.DefaultReadOptions
		if db.HasReadOptions(r) {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		contains, err := database.SetContains(cf, key, elementStr, opts)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"contains": contains,
		})
	}
}
