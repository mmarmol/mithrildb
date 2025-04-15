package handlers

import (
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "mithrildb/docs"
)

// SetupRoutes registers all HTTP routes with their handlers using a RESTful structure.
func SetupRoutes(database *db.DB, cfg *config.AppConfig, startTime time.Time) {

	http.Handle("/api/", httpSwagger.WrapHandler)

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			configGetHandler(cfg)(w, r)
		default:
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/config/update", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			configUpdateHandler(cfg)(w, r)
		default:
			respondWithNotAllowed(w)
		}
	})

	// pingHandler handles GET /ping
	//
	// @Summary      Ping endpoint
	// @Description  Returns a simple "pong" response to verify service availability.
	// @Tags         system
	// @Produce      plain
	// @Success      200  {string}  string  "pong"
	// @Failure      405  {object}  handlers.ErrorResponse  "Method not allowed"
	// @Router       /ping [get]
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("pong"))
		default:
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			metricsHandler(database, cfg, startTime)(w, r)
		default:
			respondWithNotAllowed(w)
		}
	})

	// Column families
	http.HandleFunc("/families", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listFamiliesHandler(database)(w, r)
		case http.MethodPost:
			createFamilyHandler(database)(w, r)
		default:
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/keys", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listKeysHandler(database, cfg.ReadDefaults)(w, r)
		default:
			respondWithNotAllowed(w)
		}
	})

	// Full document listing
	http.HandleFunc("/documents/bulk/get", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			bulkGetHandler(database, cfg.ReadDefaults)(w, r)
		default:
			respondWithNotAllowed(w)
		}
	})

	// Bulk operations
	http.HandleFunc("/documents/bulk/put", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			bulkPutHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			listDocumentsHandler(database, cfg.ReadDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	// Single document operations with /documents/{key}
	http.HandleFunc("/documents", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			documentGetHandler(database, cfg.ReadDefaults)(w, r)
		case http.MethodDelete:
			documentDeleteHandler(database, cfg.WriteDefaults)(w, r)
		case http.MethodPost:
			documentPutHandler(database, cfg.WriteDefaults)(w, r)
		default:
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/insert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			documentInsertHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/replace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			documentReplaceHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	// Increment counter
	http.HandleFunc("/documents/counters/delta", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			deltaCountertHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	// List operations
	http.HandleFunc("/documents/lists/push", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			listPushHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/lists/unshift", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			listUnshiftHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/lists/pop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			listPopHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/lists/shift", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			listShiftHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/lists/range", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			listRangeHandler(database, cfg.ReadDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/sets/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			setAddHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/sets/remove", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			setRemoveHandler(database, cfg.WriteDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})

	http.HandleFunc("/documents/sets/contains", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			setContainsHandler(database, cfg.ReadDefaults)(w, r)
		} else {
			respondWithNotAllowed(w)
		}
	})
}
