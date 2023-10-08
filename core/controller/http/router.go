// Package v1 implements routing paths. Each services in own file.
package http

import (
	"github.com/rhmdnrhuda/unified/config"
	"github.com/rhmdnrhuda/unified/core/repository/cache"
	"github.com/rhmdnrhuda/unified/core/repository/outbound"
	"github.com/rhmdnrhuda/unified/core/repository/postgre"
	"github.com/rhmdnrhuda/unified/pkg/postgres"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/rhmdnrhuda/unified/core/usecase"
	// Swagger docs.
	_ "github.com/rhmdnrhuda/unified/docs"
	"github.com/rhmdnrhuda/unified/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       Go Clean Template API
// @description Using a translation service as an example
// @version     1.0
// @BasePath    /api
func NewRouter(handler *gin.Engine, l logger.Interface, pg *postgres.Postgres, cfg *config.Config) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// Swagger
	swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	handler.GET("/api/swagger/*any", swaggerHandler)

	// K8s probe
	handler.GET("/api/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prometheus metrics
	handler.GET("/api/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/api")
	{
		// New Talent Routes
		NewTalentRoutes(h, usecase.NewTalentUseCase(postgre.NewTalentRepository(pg), l), l, cfg)

		NewMessageRoutes(h, usecase.NewMessageUseCase(outbound.NewVertexOutbound(cfg), outbound.NewAdaOutbound(cfg),
			cache.NewCacheCommon(), postgre.NewUserRepository(pg), l), l)
	}

}
