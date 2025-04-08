package config

import (
	"encoding/json"
	"fmt"
	"mithrildb/rocks"
	"log"
	"os"
)

// LoadConfig carga la configuración desde un archivo o usa valores por defecto
func LoadConfig() rocks.Config {
	filePath := "resources/config.json"
	var config rocks.Config

	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("No se pudo leer el archivo de configuración, usando valores por defecto: %v", err)
		config = *rocks.DefaultConfig()
	} else {
		err = json.Unmarshal(file, &config)
		if err != nil {
			log.Printf("Error al parsear el archivo de configuración, usando valores por defecto: %v", err)
			config = *rocks.DefaultConfig()
		} else {
			log.Println("Configuración cargada desde archivo con éxito.")
		}
	}

	fmt.Printf("Configuración iniciada: DBPath=%s, Port=%d\n", config.DBPath, config.Port)
	return config
}
