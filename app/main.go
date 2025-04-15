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
	// Load config from resources/config.ini
	cfg := config.LoadConfig()

	// Initialize RocksDB
	if cfg.RocksDB == nil {
		log.Fatal("❌ [Database.RocksDB] section is required in config.ini")
	}

	// Inicializa RocksDB con soporte para múltiples column families
	rocksdb, families, err := db.NewRocksDBFromConfig(*cfg.RocksDB)
	if err != nil {
		log.Fatalf("Error initializing RocksDB: %v", err)
	}

	// Crea el wrapper DB (tu estructura personalizada)
	database := db.NewDB(rocksdb, families, cfg)
	defer database.Close() // Este cierre se encarga de cerrar tanto la base como los CFs

	// Setup HTTP routes
	handlers.SetupRoutes(database, &cfg, startTime)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{Addr: addr}

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-stop
	log.Println("🧨 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}

	log.Println("✅ Server stopped cleanly.")
}
