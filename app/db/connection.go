package db

import (
	"log"

	"mithrildb/config"
	"mithrildb/rocks"
)

var (
	DB  *rocks.DB
	Cfg *config.ServerConfig
)

func InitDB(cfg config.ServerConfig) {
	var err error

	DB, err = rocks.Open(cfg.DBPath)
	if err != nil {
		log.Fatal("Error abriendo base de datos:", err)
	}
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
