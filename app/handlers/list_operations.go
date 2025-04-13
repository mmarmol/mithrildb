package handlers

import (
	"encoding/json"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"strconv"
)

type listElementRequest struct {
	Element interface{} `json:"element"`
}

func ListPushHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		cf := getCfQueryParam(r)

		var req listElementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		_, err = database.ListPush(db.ListPushOptions{
			ListOpOptions: db.ListOpOptions{
				ColumnFamily: cf,
				Key:          key,
				WriteOptions: opts,
			},
			Element: req.Element,
		})
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

func ListPopHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		cf := getCfQueryParam(r)

		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		res, err := database.ListPop(db.ListOpOptions{
			ColumnFamily: cf,
			Key:          key,
			WriteOptions: opts,
		})
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"element": res,
		})
	}
}

func ListRangeHandler(database *db.DB, defaults config.ReadOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		cf := getCfQueryParam(r)

		start, _ := strconv.Atoi(r.URL.Query().Get("start"))
		end, _ := strconv.Atoi(r.URL.Query().Get("end"))

		opts := database.DefaultReadOptions
		if db.HasReadOptions(r) {
			opts = db.BuildReadOptions(r, defaults)
			defer opts.Destroy()
		}

		res, err := database.ListRange(cf, key, start, end, opts)
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"list": res,
		})
	}
}

func ListUnshiftHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		cf := getCfQueryParam(r)

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

		_, err = database.ListUnshift(db.ListPushOptions{
			ListOpOptions: db.ListOpOptions{
				ColumnFamily: cf,
				Key:          key,
				WriteOptions: opts,
			},
			Element: req.Element,
		})
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

func ListShiftHandler(database *db.DB, defaults config.WriteOptionsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := getQueryParam(r, "key")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		cf := getCfQueryParam(r)

		opts := database.DefaultWriteOptions
		if db.HasWriteOptions(r) {
			opts = db.BuildWriteOptions(r, defaults)
			defer opts.Destroy()
		}

		result, err := database.ListShift(db.ListOpOptions{
			ColumnFamily: cf,
			Key:          key,
			WriteOptions: opts,
		})
		if err != nil {
			mapAndRespondWithError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"element": result,
		})
	}
}
