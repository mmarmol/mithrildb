package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/ini.v1"
)

type ServerConfig struct {
	Port int
}

type RocksDBConfig struct {
	DBPath            string
	CreateIfMissing   bool
	WriteBufferSize   int
	MaxWriteBufferNum int
	BlockCacheSize    int
	StatsDumpPeriod   time.Duration
	MaxOpenFiles      int
	EnableCompression bool
	CompressionType   string
}

type WriteOptionsConfig struct {
	Sync       bool
	DisableWAL bool
	NoSlowdown bool
}

type ReadOptionsConfig struct {
	FillCache bool
	ReadTier  string // "all", "cache-only"
}

type AppConfig struct {
	Server        ServerConfig
	RocksDB       *RocksDBConfig
	WriteDefaults WriteOptionsConfig
	ReadDefaults  ReadOptionsConfig
}

type UpdateResult struct {
	Applied  []string
	Pending  []string
	Rejected map[string]string
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
				StatsDumpPeriod:   r.Key("StatsDumpPeriod").MustDuration(1 * time.Minute),
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
	}

	// Si no se cargó desde archivo, establecer defaults manualmente
	if cfg.RocksDB == nil {
		cfg.RocksDB = &RocksDBConfig{
			DBPath:            "./data",
			CreateIfMissing:   true,
			WriteBufferSize:   32 * 1024 * 1024,
			MaxWriteBufferNum: 2,
			BlockCacheSize:    128 * 1024 * 1024,
			StatsDumpPeriod:   1 * time.Minute,
			MaxOpenFiles:      500,
			EnableCompression: false,
			CompressionType:   "snappy",
		}
	}

	fmt.Printf("✅ Configuración cargada:\n%+v\n", cfg)
	return cfg
}

// UpdateConfigFromMap processes updates to the RocksDB section of the config file.
func UpdateConfigFromMap(cfg *AppConfig, configPath string, req map[string]interface{}) (*UpdateResult, error) {
	iniFile, err := ini.Load(configPath)
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
		if err := iniFile.SaveTo(configPath); err != nil {
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
