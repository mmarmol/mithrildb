package config

import (
	"fmt"
	"log"
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

// LoadConfig lee y aplica la configuración desde el archivo resources/config.ini
func LoadConfig() AppConfig {
	cfg := AppConfig{
		Server: ServerConfig{
			Port: 5126,
		},
	}

	file, err := ini.Load("resources/config.ini")
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
