package cmd

import (
	"encoding/base64"
	"field-service/clients"
	"field-service/common/gcs"
	"field-service/common/response"
	"field-service/config"
	"field-service/constants"
	"field-service/controllers"
	"field-service/domain/models"
	"field-service/middlewares"
	"field-service/repositories"
	"field-service/routes"
	"field-service/services"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()
		config.Init()
		db, err := config.InitDatabase()
		if err != nil {
			panic(err)
		}
		loc, err := time.LoadLocation("Asia/Jakarta")

		if err != nil {
			panic(err)
		}
		time.Local = loc

		err = db.AutoMigrate(
			&models.Field{}, &models.FieldSchedule{}, &models.Time{},
		)
		if err != nil {
			panic(err)
		}

		gcs := initGCS()
		client := clients.NewClientRegistry()
		repository := repositories.NewRepositoryRegistry(db)
		service := services.NewServiceRegistry(repository, gcs)

		controller := controllers.NewControllerRegistry(service)

		router := gin.Default()
		router.Use(middlewares.HandlePanic())

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

		port := fmt.Sprintf(":%d", config.Config.Port)
		err = router.Run(port)
		if err != nil {
			panic(err)
		}
	},
}

func Run() {
	err := command.Execute()
	if err != nil {
		panic(err)
	}
}

func initGCS() gcs.IGSClient {
	decode, err := base64.StdEncoding.DecodeString(config.Config.GCSPrivateKey)
	if err != nil {
		panic(err)
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
