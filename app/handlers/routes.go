package handlers

import (
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"strings"
	"time"
)

// SetupRoutes registers all HTTP routes with their handlers using a RESTful structure.
func SetupRoutes(database *db.DB, cfg config.AppConfig, startTime time.Time) {

	// Config endpoints
	http.HandleFunc("/config", ConfigGetHandler(cfg))
	http.HandleFunc("/config/update", ConfigUpdateHandler(cfg, database))

	// Standard endpoints
	http.HandleFunc("/ping", PingHandler())
	http.HandleFunc("/metrics", MetricsHandler(database, cfg, startTime))

	// Column families
	http.HandleFunc("/families", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ListFamiliesHandler(database)(w, r)
		case http.MethodPost:
			CreateFamilyHandler(database)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Key-only listing
	http.HandleFunc("/keys", ListKeysHandler(database, cfg.ReadDefaults))

	// Full document listing
	http.HandleFunc("/documents", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ListDocumentsHandler(database, cfg.ReadDefaults)(w, r)
		case http.MethodPost:
			PutHandler(database, cfg.WriteDefaults)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Bulk operations
	http.HandleFunc("/documents/bulk", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			MultiPutHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			MultiGetHandler(database, cfg.ReadDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Single document operations with /documents/{key}
	http.HandleFunc("/documents/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/documents/")
		if key == "" {
			http.Error(w, "missing document key", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			GetHandler(database, cfg.ReadDefaults, key)(w, r)
		case http.MethodDelete:
			DeleteHandler(database, cfg.WriteDefaults, key)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/insert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			InsertHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/replace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ReplaceHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Increment counter
	http.HandleFunc("/documents/counters/delta", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			DeltaCountertHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// List operations
	http.HandleFunc("/documents/lists/push", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ListPushHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/lists/unshift", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ListUnshiftHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/lists/pop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ListPopHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/lists/shift", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ListShiftHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/lists/range", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ListRangeHandler(database, cfg.ReadDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/sets/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			SetAddHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/sets/remove", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			SetRemoveHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/documents/sets/contains", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			SetContainsHandler(database, cfg.ReadDefaults)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
