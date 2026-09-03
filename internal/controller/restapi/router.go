package restapi

import (
	"net/http"

	"github.com/evrone/go-clean-template/config"
	_ "github.com/evrone/go-clean-template/docs" // Swagger docs.
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	v1 "github.com/evrone/go-clean-template/internal/controller/restapi/v1"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// NewRouter -.
// Swagger spec:
//
//	@title       Go Clean Template API
//	@description Multi-domain clean architecture template with translation, user, and task management
//	@version     1.0
//	@host        localhost:8080
//	@BasePath    /v1
//	@securityDefinitions.apikey BearerAuth
//	@in header
//	@name Authorization
func NewRouter(app *gin.Engine, cfg *config.Config, t usecase.Translation, u usecase.User, tk usecase.Task, jwtManager *jwt.Manager, l logger.Interface) {
	// Options
	app.Use(middleware.Logger(l))
	app.Use(middleware.Recovery(l))

	// Prometheus metrics
	if cfg.Metrics.Enabled {
		prometheus := middleware.NewMetrics("my-service-name")
		app.GET("/metrics", prometheus.Handler())
		app.Use(prometheus.Middleware())
	}

	// Swagger
	if cfg.Swagger.Enabled {
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// K8s probe
	app.GET("/healthz", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	// Routers
	apiV1Group := app.Group("/v1")
	{
		if cfg.Tracing.Enabled {
			apiV1Group.Use(otelgin.Middleware(cfg.App.Name))
		}

		v1.NewRoutes(apiV1Group, t, u, tk, jwtManager, l)
	}
}
