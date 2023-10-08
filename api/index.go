package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/log"
	"github.com/rhmdnrhuda/unified/config"
	main_app "github.com/rhmdnrhuda/unified/core/app"
	"github.com/rhmdnrhuda/unified/pkg/logger"
	"github.com/rhmdnrhuda/unified/pkg/postgres"
	"net/http"
)

var (
	app *gin.Engine
	cfg *config.Config
	err error
	pg  *postgres.Postgres
)

func Handler(w http.ResponseWriter, r *http.Request) {
	cfg, err = config.NewConfig()
	if err != nil {
		log.Error("Config error: %s", err)
	}

	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err = postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("http - Run - postgres.New: %w", err))
	}

	defer pg.Close()

	app = main_app.ProvideHandler(pg, l, cfg)

	app.ServeHTTP(w, r)
}
