package main

import (
	"log"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/handlers"
	"net/http"
	"strconv"
)

func main() {
	cfg := config.LoadConfig()

	db.InitDB(cfg)
	defer db.CloseDB()

	handlers.SetupRoutes()

	address := ":" + strconv.Itoa(cfg.Port)
	log.Printf("Servidor escuchando en %s\n", address)
	log.Fatal(http.ListenAndServe(address, nil))

}
