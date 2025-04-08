package config

import (
	"fmt"
	"log"

	"gopkg.in/ini.v1"
)

type ServerConfig struct {
	DBPath string
	Port   int
}

func LoadConfig() ServerConfig {
	cfg := ServerConfig{
		DBPath: "/data/db", // Valores por defecto
		Port:   5126,
	}

	file, err := ini.Load("resources/config.ini")
	if err != nil {
		log.Printf("Error cargando config.ini, utilizando valores por defecto: %v", err)
		return cfg
	}

	cfg.DBPath = file.Section("Server").Key("DBPath").MustString(cfg.DBPath)
	cfg.Port = file.Section("Server").Key("Port").MustInt(cfg.Port)

	fmt.Printf("Configuración iniciada: DBPath=%s, Port=%d\n", cfg.DBPath, cfg.Port)
	return cfg
}
