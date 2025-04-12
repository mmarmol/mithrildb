package main

import (
	"context"
	"fmt"
	"log"
	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/handlers"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var startTime = time.Now()

func main() {
	cfg := config.LoadConfig()

	database, err := db.Open(db.RocksDBOptions{
		DBPath: cfg.DBPath,
	})
	if err != nil {
		log.Fatalf("Error al abrir la base de datos: %v", err)
	}
	defer database.Close()

	handlers.SetupRoutes(database, cfg.DBPath, startTime) // ahora le pasamos el DBPath para el endpoint /stats

	// Configura el servidor HTTP
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{Addr: addr}

	// Canal para señales del sistema
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Ejecuta el servidor en una goroutine
	go func() {
		log.Printf("Servidor escuchando en %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
		}
	}()

	// Espera señal para detener
	<-stop
	log.Println("Deteniendo servidor...")

	// Apaga el servidor de forma controlada
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error durante el apagado del servidor: %v", err)
	}

	log.Println("Servidor detenido correctamente.")
}
