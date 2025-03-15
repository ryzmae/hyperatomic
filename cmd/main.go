package main

import (
	"fmt"

	"github.com/ryzmae/hyperatomic/internal/config"
	"github.com/ryzmae/hyperatomic/internal/logger"
	"github.com/ryzmae/hyperatomic/internal/server"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("❌ Failed to load config:", err)
		return
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg)
	if err != nil {
		fmt.Println("❌ Failed to initialize logger:", err)
		return
	}
	defer log.Close()

	// Start TCP server
	srv, err := server.NewServer(cfg, log)
	if err != nil {
		fmt.Println("❌ Failed to start server:", err)
		return
	}

	log.Info("🚀 Server started on port %d", cfg.TCP.Port)
	srv.HandleConnections()
}
