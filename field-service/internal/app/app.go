package app

import (
	"context"
	"encoding/base64"
	"field-service/clients"
	"field-service/common/gcs"
	"field-service/common/response"
	"field-service/config"
	"field-service/constants"
	"field-service/controllers"
	healthController "field-service/controllers/health"
	_ "field-service/docs"
	"field-service/domain/models"
	"field-service/middlewares"
	"field-service/repositories"
	"field-service/routes"
	healthRoute "field-service/routes/health"
	"field-service/services"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

type App struct {
	db     *gorm.DB
	server *http.Server
}

func New() (*App, error) {
	if err := loadConfig(); err != nil {
		return nil, err
	}

	db, err := config.InitDatabase()
	if err != nil {
		return nil, err
	}

	if err := setTimezone("Asia/Jakarta"); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	router := buildRouter(db)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.Port),
		Handler: router,
	}

	return &App{db: db, server: server}, nil
}

func (a *App) Run() error {
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	a.closeDB()
	logrus.Info("Server exiting")
	return nil
}

func loadConfig() error {
	_ = godotenv.Load()
	config.Init()
	if err := config.Config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	return nil
}

func setTimezone(name string) error {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Field{}, &models.FieldSchedule{}, &models.Time{},
	)
}

func buildRouter(db *gorm.DB) *gin.Engine {
	gcsClient := initGCS()
	client := clients.NewClientRegistry()
	repository := repositories.NewRepositoryRegistry(db)
	service := services.NewServiceRegistry(repository, gcsClient)
	controller := controllers.NewControllerRegistry(service)

	router := gin.Default()
	registerGlobalMiddleware(router)
	registerRootRoutes(router)
	registerInlineCORS(router)
	registerHealthRoutes(router, db)
	registerRateLimiter(router)
	registerAPIRoutes(router, controller, client)
	registerSwagger(router)

	return router
}

func registerGlobalMiddleware(router *gin.Engine) {
	router.Use(middlewares.RequestLogger())
	router.Use(middlewares.HandlePanic())
	router.Use(middlewares.SecurityHeaders())
	router.Use(middlewares.CORS())
}

func registerRootRoutes(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, response.Response{
			Status:  constants.Error,
			Message: fmt.Sprintf("Path %s", http.StatusText(http.StatusNotFound)),
		})
	})
	router.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, response.Response{
			Status:  constants.Success,
			Message: "Welcome to Field Service",
		})
	})
}

func registerInlineCORS(router *gin.Engine) {
	router.Use(func(context *gin.Context) {
		context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		context.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-apikey, x-request-at")
		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(204)
			return
		}
		context.Next()
	})
}

func registerHealthRoutes(router *gin.Engine, db *gorm.DB) {
	hc := healthController.NewHealthController(db)
	hr := healthRoute.NewHealthRoute(router, hc)
	hr.Serve()
}

func registerRateLimiter(router *gin.Engine) {
	lmt := tollbooth.NewLimiter(
		config.Config.RateLimiterMaxRequest,
		&limiter.ExpirableOptions{
			DefaultExpirationTTL: time.Duration(config.Config.RateLimiterTimeSecond) * time.Second,
		},
	)
	router.Use(middlewares.RateLimiter(lmt))
}

func registerAPIRoutes(router *gin.Engine, controller controllers.IControllerRegistry, client clients.IClientRegistry) {
	group := router.Group("/api/v1")
	route := routes.NewRouteRegistry(group, controller, client)
	route.Serve()
}

func registerSwagger(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func initGCS() gcs.IGSClient {
	if config.Config.GCSPrivateKey == "" {
		logrus.Warn("GCSPrivateKey is empty, using MockGSClient")
		return gcs.NewGSClient(gcs.ServiceAccountKeyJson{}, "")
	}

	decode, err := base64.StdEncoding.DecodeString(config.Config.GCSPrivateKey)
	if err != nil {
		logrus.Errorf("failed to decode GCSPrivateKey: %v", err)
		logrus.Warn("falling back to MockGSClient due to invalid key")
		return gcs.NewGSClient(gcs.ServiceAccountKeyJson{}, "")
	}

	privateKeyPEM := strings.ReplaceAll(string(decode), `\n`, "\n")
	gcsServiceAccount := gcs.ServiceAccountKeyJson{
		Type:                    config.Config.GCSType,
		ProjectID:               config.Config.GCSProjectID,
		PrivateKeyID:            config.Config.GCSPrivateKeyID,
		PrivateKey:              privateKeyPEM,
		ClientEmail:             config.Config.GCSClientEmail,
		ClientID:                config.Config.GCSClientID,
		AuthURI:                 config.Config.GCSAuthURI,
		TokenURI:                config.Config.GCSTokenURI,
		AuthProviderX509CertURL: config.Config.GCSAuthProviderX509CertURL,
		ClientX509CertURL:       config.Config.GCSClientX509CertURL,
		UniverseDomain:          config.Config.GCSUniverseDomain,
	}
	return gcs.NewGSClient(gcsServiceAccount, config.Config.GCSBucketName)
}

func (a *App) closeDB() {
	sqlDB, _ := a.db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}
