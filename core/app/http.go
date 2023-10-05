// Package http configures and runs application.
package app

import (
	"fmt"
	v1 "github.com/temukan-co/monolith/core/controller/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/temukan-co/monolith/config"
	"github.com/temukan-co/monolith/pkg/httpserver"
	"github.com/temukan-co/monolith/pkg/logger"
	"github.com/temukan-co/monolith/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("http - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	handler := ProvideHandler(pg, l, cfg)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("http - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("http - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("http - Run - httpServer.Shutdown: %w", err))
	}
}

func ProvideHandler(pg *postgres.Postgres, l *logger.Logger, cfg *config.Config) *gin.Engine {
	// HTTP Server
	handler := gin.New()
	handler.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowAllOrigins:  true,
		AllowCredentials: true,
	}))

	v1.NewRouter(handler, l, pg, cfg)

	return handler
}
