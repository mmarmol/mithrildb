package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/ini.v1"
)

// ServerConfig represents the HTTP server configuration.
// @Description Listening port configuration for the REST API.
type ServerConfig struct {
	// Port where the HTTP server listens
	Port int `json:"port"`
}

// RocksDBConfig holds configuration parameters for the RocksDB database.
// @Description Detailed configuration for the RocksDB storage backend.
type RocksDBConfig struct {
	DBPath            string `json:"db_path"`              // Path to the data directory
	CreateIfMissing   bool   `json:"create_if_missing"`    // Whether to create the DB if it doesn't exist
	WriteBufferSize   int    `json:"write_buffer_size"`    // Size of the write buffer in bytes
	MaxWriteBufferNum int    `json:"max_write_buffer_num"` // Maximum number of write buffers
	BlockCacheSize    int    `json:"block_cache_size"`     // Size of the block cache in bytes
	StatsDumpPeriod   string `json:"stats_dump_period"`    // Frequency for dumping RocksDB statistics (e.g. "30s", "1m")
	MaxOpenFiles      int    `json:"max_open_files"`       // Maximum number of open files
	EnableCompression bool   `json:"enable_compression"`   // Whether compression is enabled
	CompressionType   string `json:"compression_type"`     // Compression type: snappy, zstd, lz4, none
}

// WriteOptionsConfig represents configurable write options for RocksDB.
// @Description Controls the write behavior in RocksDB.
type WriteOptionsConfig struct {
	Sync       bool `json:"sync"`        // Wait for disk sync on write
	DisableWAL bool `json:"disable_wal"` // Disable the Write-Ahead Log
	NoSlowdown bool `json:"no_slowdown"` // Avoid blocking if RocksDB is busy
}

// ReadOptionsConfig represents configurable read options.
// @Description Controls how data is read from the database.
type ReadOptionsConfig struct {
	FillCache bool   `json:"fill_cache"` // Whether to fill cache on reads
	ReadTier  string `json:"read_tier"`  // Read level: all or cache-only
}

// AppConfig represents the full configuration loaded by the application.
// @Description Global server and database configuration.
type AppConfig struct {
	Server        ServerConfig       `json:"server"`
	RocksDB       *RocksDBConfig     `json:"rocksdb"`
	WriteDefaults WriteOptionsConfig `json:"write_defaults"`
	ReadDefaults  ReadOptionsConfig  `json:"read_defaults"`
	Expiration    ExpirationConfig   `json:"expiration"`
}

// UpdateResult represents the result of a configuration update.
// @Description Outcome when applying changes to the configuration file.
type UpdateResult struct {
	Applied  []string          `json:"applied"`          // Parameters applied immediately
	Pending  []string          `json:"requires_restart"` // Parameters that require a restart
	Rejected map[string]string `json:"rejected"`         // Parameters rejected with reasons
}

// ExpirationConfig holds configuration for the expiration background service.
//
// @Description Expiration (TTL) service configuration.
type ExpirationConfig struct {
	TickInterval       string  `json:"tick_interval"`        // Interval between expiration runs (e.g. "1m", "30s")
	MaxPerCycle        int     `json:"max_per_cycle"`        // Maximum documents processed per cycle
	AutoScale          bool    `json:"auto_scale"`           // Enable or disable automatic scaling
	MinPerCycle        int     `json:"min_per_cycle"`        // Minimum limit when scaling down
	MaxPerCycleLimit   int     `json:"max_per_cycle_limit"`  // Maximum limit when scaling up
	ScaleStep          int     `json:"scale_step"`           // Step amount to increase/decrease
	ScaleDownThreshold float64 `json:"scale_down_threshold"` // Scale down if deletions < MaxPerCycle * threshold
	ScaleUpFactor      float64 `json:"scale_up_factor"`      // Scale up if duration < TickInterval * factor
}

func LoadConfig() AppConfig {
	cfg := AppConfig{
		Server: ServerConfig{
			Port: 5126,
		},
	}
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "resources/config.ini"
	}

	file, err := ini.Load(path)
	if err != nil {
		log.Printf("⚠️  Error cargando config.ini, usando valores por defecto: %v", err)
	} else {
		// [Server]
		s := file.Section("Server")
		cfg.Server.Port = s.Key("Port").MustInt(cfg.Server.Port)

		// [Database.RocksDB]
		if file.HasSection("Database.RocksDB") {
			r := file.Section("Database.RocksDB")
			cfg.RocksDB = &RocksDBConfig{
				DBPath:            r.Key("DBPath").MustString("./data"),
				CreateIfMissing:   r.Key("CreateIfMissing").MustBool(true),
				WriteBufferSize:   r.Key("WriteBufferSize").MustInt(32 * 1024 * 1024), // 32MB
				MaxWriteBufferNum: r.Key("MaxWriteBufferNumber").MustInt(2),
				BlockCacheSize:    r.Key("BlockCacheSize").MustInt(128 * 1024 * 1024), // 128MB
				StatsDumpPeriod:   r.Key("StatsDumpPeriod").MustString("1m"),
				MaxOpenFiles:      r.Key("MaxOpenFiles").MustInt(500),
				EnableCompression: r.Key("EnableCompression").MustBool(false),
				CompressionType:   r.Key("CompressionType").MustString("snappy"),
			}
		} else {
			log.Println("⚠️  No se encontró [Database.RocksDB] en config.ini — se usarán valores por defecto.")
		}
		// [WriteOptions]
		wo := file.Section("Database.RocksDB.WriteOptions")
		cfg.WriteDefaults = WriteOptionsConfig{
			Sync:       wo.Key("Sync").MustBool(false),
			DisableWAL: wo.Key("DisableWAL").MustBool(false),
			NoSlowdown: wo.Key("NoSlowdown").MustBool(false),
		}
		// [ReadOptions]
		ro := file.Section("Database.RocksDB.ReadOptions")
		cfg.ReadDefaults = ReadOptionsConfig{
			FillCache: ro.Key("FillCache").MustBool(true),
			ReadTier:  ro.Key("ReadTier").MustString("all"),
		}
		// [Expiration]
		exp := file.Section("Expiration")
		cfg.Expiration = ExpirationConfig{
			TickInterval:       exp.Key("TickInterval").MustString("1m"),
			MaxPerCycle:        exp.Key("MaxPerCycle").MustInt(500),
			AutoScale:          exp.Key("AutoScale").MustBool(true),
			MinPerCycle:        exp.Key("MinPerCycle").MustInt(100),
			MaxPerCycleLimit:   exp.Key("MaxPerCycleLimit").MustInt(5000),
			ScaleStep:          exp.Key("ScaleStep").MustInt(100),
			ScaleDownThreshold: exp.Key("ScaleDownThreshold").MustFloat64(0.25),
			ScaleUpFactor:      exp.Key("ScaleUpFactor").MustFloat64(0.5),
		}
	}

	// if file not loaded them defaults
	if cfg.RocksDB == nil {
		cfg.RocksDB = &RocksDBConfig{
			DBPath:            "./data",
			CreateIfMissing:   true,
			WriteBufferSize:   32 * 1024 * 1024,
			MaxWriteBufferNum: 2,
			BlockCacheSize:    128 * 1024 * 1024,
			StatsDumpPeriod:   "1m",
			MaxOpenFiles:      500,
			EnableCompression: false,
			CompressionType:   "snappy",
		}
	}

	fmt.Printf("✅ Configuración cargada:\n%+v\n", cfg)
	return cfg
}

// UpdateConfigFromMap processes updates to the RocksDB section of the config file.
func UpdateConfigFromMap(cfg *AppConfig, req map[string]interface{}) (*UpdateResult, error) {
	iniFile, err := ini.Load(cfg.RocksDB.DBPath)
	if err != nil {
		return nil, err
	}

	dbSection := iniFile.Section("Database.RocksDB")
	applied := []string{}
	pending := []string{}
	rejected := make(map[string]string)
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
		if err := iniFile.SaveTo(cfg.RocksDB.DBPath); err != nil {
			return nil, err
		}
	}

	return &UpdateResult{
		Applied:  applied,
		Pending:  pending,
		Rejected: rejected,
	}, nil
}

// Helper: convert interface{} to int if possible
func toInt(val interface{}) (int, bool) {
	switch v := val.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	case string:
		var i int
		_, err := fmt.Sscanf(v, "%d", &i)
		return i, err == nil
	default:
		return 0, false
	}
}

// Helper: check if a duration string is valid
func isValidDuration(dur string) bool {
	_, err := time.ParseDuration(dur)
	return err == nil
}

// Helper: validate compression type
func isValidCompression(value string) bool {
	switch value {
	case "snappy", "zstd", "lz4", "none":
		return true
	default:
		return false
	}
}
