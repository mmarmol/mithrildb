// handlers/config_update.go
package handlers

import (
	"encoding/json"
	"fmt"
	"mithrildb/config"
	"mithrildb/db"
	"net/http"
	"reflect"
	"time"

	"gopkg.in/ini.v1"
)

type configUpdateRequest map[string]interface{}

type updateResponse struct {
	Applied  []string          `json:"applied"`
	Pending  []string          `json:"requires_restart"`
	Rejected map[string]string `json:"rejected"`
}

// ConfigUpdateHandler handles POST /config/update
func ConfigUpdateHandler(cfg *config.AppConfig, dbInstance *db.DB, configPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req configUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithErrInvalidJSONBody(w)
			return
		}

		applied := []string{}
		pending := []string{}
		rejected := map[string]string{}

		iniFile, err := ini.Load(configPath)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not read config file")
			return
		}

		dbSection := iniFile.Section("Database.RocksDB")
		modified := false

		for key, value := range req {
			switch key {
			case "WriteBufferSize", "MaxWriteBufferNumber", "BlockCacheSize", "MaxOpenFiles":
				intVal, ok := toInt(value)
				if !ok || intVal < 0 {
					rejected[key] = "must be a non-negative integer"
					continue
				}
				dbSection.Key(key).SetValue(fmt.Sprint(intVal))
				applied = append(applied, key)
				modified = true
				// optionally: apply to db in memory if supported
			case "StatsDumpPeriod":
				strVal, ok := value.(string)
				if !ok || !isValidDuration(strVal) {
					rejected[key] = "must be a valid duration like '30s', '1m'"
					continue
				}
				dbSection.Key(key).SetValue(strVal)
				applied = append(applied, key)
				modified = true
			case "CompressionType":
				strVal, ok := value.(string)
				if !ok || !isValidCompression(strVal) {
					rejected[key] = "must be one of: snappy, zstd, lz4, none"
					continue
				}
				dbSection.Key(key).SetValue(strVal)
				pending = append(pending, key)
				modified = true
			default:
				rejected[key] = "unsupported or read-only field"
			}
		}

		if modified {
			_ = iniFile.SaveTo(configPath)
		}

		resp := updateResponse{
			Applied:  applied,
			Pending:  pending,
			Rejected: rejected,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func toInt(val interface{}) (int, bool) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Float64:
		return int(v.Float()), true
	case reflect.Int, reflect.Int64:
		return int(v.Int()), true
	case reflect.String:
		var i int
		_, err := fmt.Sscanf(v.String(), "%d", &i)
		return i, err == nil
	default:
		return 0, false
	}
}

func isValidDuration(dur string) bool {
	_, err := time.ParseDuration(dur)
	return err == nil
}

func isValidCompression(value string) bool {
	switch value {
	case "snappy", "zstd", "lz4", "none":
		return true
	default:
		return false
	}
}
