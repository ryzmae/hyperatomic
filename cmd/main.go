package main

import (
	"fmt"

	"github.com/ryzmae/hyperatomic/internal/config"
	"github.com/ryzmae/hyperatomic/internal/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	log, err := logger.NewLogger(cfg)
	if err != nil {
		fmt.Println("Failed to initialize logger:", err)
		return
	}

	log.Info("HyperAtomic started successfully!")
	log.Debug("Config: %+v", cfg)
	log.Close()
}
