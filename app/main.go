package main

import (
	"context"
	"fmt"
	"log"
	"mithrildb/bootstrap"
	"mithrildb/config"
	"mithrildb/handlers"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var startTime = time.Now()

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database (RocksDB + CFs)
	database := bootstrap.InitDatabase(&cfg)
	defer database.Close()

	// Setup event system (queue, fanout, TTL listener)
	bootstrap.InitEventSystem(database)

	// Setup expiration service (cron + stats)
	expirer := bootstrap.InitExpirationService(database, cfg)

	// Setup HTTP routes
	handlers.SetupRoutes(database, expirer, &cfg, startTime)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{Addr: addr}

	// Graceful shutdown handling
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
