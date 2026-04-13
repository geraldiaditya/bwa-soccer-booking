package cmd

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
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "field-service/docs"
)

// @title Field Service API
// @version 1.0
// @description Microservice for managing soccer fields and schedules.
// @host localhost:8002
// @BasePath /api/v1

var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()
		config.Init()
	if err := config.Config.Validate(); err != nil {
		logrus.Fatalf("invalid configuration: %v", err)
	}
		db, err := config.InitDatabase()
		if err != nil {
			logrus.Fatal(err)
		}
		loc, err := time.LoadLocation("Asia/Jakarta")

		if err != nil {
			logrus.Fatal(err)
		}
		time.Local = loc

		err = db.AutoMigrate(
			&models.Field{}, &models.FieldSchedule{}, &models.Time{},
		)
		if err != nil {
			logrus.Fatal(err)
		}

		gcs := initGCS()
		client := clients.NewClientRegistry()
		repository := repositories.NewRepositoryRegistry(db)
		service := services.NewServiceRegistry(repository, gcs)

		controller := controllers.NewControllerRegistry(service)

		router := gin.Default()
		router.Use(middlewares.RequestLogger())
		router.Use(middlewares.HandlePanic())
		router.Use(middlewares.SecurityHeaders())
		router.Use(middlewares.CORS())

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

		hc := healthController.NewHealthController(db)
		hr := healthRoute.NewHealthRoute(router, hc)
		hr.Serve()

		lmt := tollbooth.NewLimiter(
			config.Config.RateLimiterMaxRequest,
			&limiter.ExpirableOptions{
				DefaultExpirationTTL: time.Duration(config.Config.RateLimiterTimeSecond) * time.Second,
			},
		)
		router.Use(middlewares.RateLimiter(lmt))

		group := router.Group("/api/v1")
		route := routes.NewRouteRegistry(group, controller, client)
		route.Serve()

		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		port := fmt.Sprintf(":%d", config.Config.Port)
		srv := &http.Server{
			Addr:    port,
			Handler: router,
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("listen: %s\n", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logrus.Info("Shutdown Server ...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logrus.Fatal("Server Shutdown:", err)
		}

		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}

		logrus.Info("Server exiting")
	},
}

func Run() {
	err := command.Execute()
	if err != nil {
		logrus.Fatal(err)
	}
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
	gcsClient := gcs.NewGSClient(
		gcsServiceAccount,
		config.Config.GCSBucketName)
	return gcsClient
}
