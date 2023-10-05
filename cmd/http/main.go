package main

import (
	"log"

	"github.com/temukan-co/monolith/config"
	"github.com/temukan-co/monolith/core/app"
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
