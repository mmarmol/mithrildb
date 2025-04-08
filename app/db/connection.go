package db

import (
	"log"

	"mithrildb/rocks"
)

var (
	DB  *rocks.DB
	Cfg *rocks.Config
)

func InitDB(cfg rocks.Config) {
	var err error

	// Usando la configuración del paquete rocks
	Cfg := &rocks.Config{
		CreateIfMissing:                cfg.CreateIfMissing,
		ErrorIfExists:                  cfg.ErrorIfExists,
		ParanoidChecks:                 cfg.ParanoidChecks,
		WriteBufferSize:                cfg.WriteBufferSize,
		MaxWriteBufferNumber:           cfg.MaxWriteBufferNumber,
		MinWriteBufferNumberToMerge:    cfg.MinWriteBufferNumberToMerge,
		MaxBackgroundCompactions:       cfg.MaxBackgroundCompactions,
		MaxBackgroundFlushes:           cfg.MaxBackgroundFlushes,
		Compression:                    cfg.Compression,
		CompactionStyle:                cfg.CompactionStyle,
		Level0FileNumCompactionTrigger: cfg.Level0FileNumCompactionTrigger,
		Level0SlowdownWritesTrigger:    cfg.Level0SlowdownWritesTrigger,
		Level0StopWritesTrigger:        cfg.Level0StopWritesTrigger,
		MaxCompactionBytes:             cfg.MaxCompactionBytes,
		BlockSize:                      cfg.BlockSize,
		BlockCacheSize:                 cfg.BlockCacheSize,
		BlockCacheCompressedSize:       cfg.BlockCacheCompressedSize,
		BloomFilterPolicy:              cfg.BloomFilterPolicy,
		WriteSync:                      cfg.WriteSync,
		DisableAutoCompactions:         cfg.DisableAutoCompactions,
		MaxOpenFiles:                   cfg.MaxOpenFiles,
		MaxLogFileSize:                 cfg.MaxLogFileSize,
		LogFileTimeToRoll:              cfg.LogFileTimeToRoll,
		ReuseLogs:                      cfg.ReuseLogs,
		Statistics:                     cfg.Statistics,
		DisableMemtableActivity:        cfg.DisableMemtableActivity,
		MaxTotalWALSize:                cfg.MaxTotalWALSize,
		WalTTLSeconds:                  cfg.WalTTLSeconds,
		WalSizeLimitMB:                 cfg.WalSizeLimitMB,
		ManifestPreallocationSize:      cfg.ManifestPreallocationSize,
		DBPath:                         cfg.DBPath,
		Port:                           cfg.Port,
	}

	DB, err = rocks.Open(cfg.DBPath, Cfg)
	if err != nil {
		log.Fatal("Error abriendo base de datos:", err)
	}
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
