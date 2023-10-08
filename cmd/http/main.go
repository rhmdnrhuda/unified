package main

import (
	"log"

	"github.com/rhmdnrhuda/unified/config"
	"github.com/rhmdnrhuda/unified/core/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
