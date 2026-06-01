package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"payment-service/clients"
	midtransClient "payment-service/clients/midtrans"
	"payment-service/common/gcs"
	"payment-service/common/response"
	"payment-service/config"
	"payment-service/constants"
	controllers "payment-service/controllers/http"
	kafkaClient "payment-service/controllers/kafka"
	"payment-service/domain/models"
	"payment-service/middlewares"
	"payment-service/repositories"
	"payment-service/routes"
	"payment-service/services"
	"strings"
	"syscall"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const shutdownTimeout = 10 * time.Second

type dependencies struct {
	Controller controllers.IControllerRegistry
	Client     clients.IClientRegistry
}

func Run(ctx context.Context) error {
	_ = godotenv.Load()
	config.Init()

	db, err := config.InitDatabase()
	if err != nil {
		return err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return err
	}
	time.Local = loc

	if err = db.AutoMigrate(
		&models.Payment{},
		&models.PaymentHistory{},
	); err != nil {
		return err
	}

	gcsClient, err := initGCS(config.Config)
	if err != nil {
		return err
	}
	kafka, err := kafkaClient.NewKafkaRegistry(config.Config.Kafka.Brokers, config.Config.Kafka.MaxRetry)
	if err != nil {
		return err
	}
	defer func() {
		_ = kafka.Close()
	}()
	midtrans := midtransClient.NewMidTransClient(
		config.Config.Midtrans.ServerKey,
		config.Config.Midtrans.IsProduction,
	)
	client := clients.NewClientRegistry()
	repository := repositories.NewRepositoryRegistry(db)
	service := services.NewServiceRegistry(repository, gcsClient, kafka, midtrans)
	controller := controllers.NewControllerRegistry(service)

	router := newRouter(config.Config, dependencies{
		Controller: controller,
		Client:     client,
	})

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.Port),
		Handler: router,
	}

	return serve(ctx, server)
}

func newRouter(appConfig config.AppConfig, deps dependencies) *gin.Engine {
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
			Message: "Welcome to Payment Service",
		})
	})

	router.Use(func(context *gin.Context) {
		context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		context.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-apikey, x-request-at")
		if context.Request.Method == http.MethodOptions {
			context.AbortWithStatus(http.StatusNoContent)
			return
		}
		context.Next()
	})

	lmt := tollbooth.NewLimiter(
		appConfig.RateLimiterMaxRequest,
		&limiter.ExpirableOptions{
			DefaultExpirationTTL: time.Duration(appConfig.RateLimiterTimeSecond) * time.Second,
		},
	)
	router.Use(middlewares.RateLimiter(lmt))

	group := router.Group("/api/v1")
	route := routes.NewRouteRegistry(group, deps.Controller, deps.Client)
	route.Serve()

	return router
}

func serve(ctx context.Context, server *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-errCh
	}
}

func initGCS(appConfig config.AppConfig) (gcs.IGSClient, error) {
	decode, err := base64.StdEncoding.DecodeString(appConfig.GCSPrivateKey)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := strings.ReplaceAll(string(decode), `\n`, "\n")
	gcsServiceAccount := gcs.ServiceAccountKeyJson{
		Type:                    appConfig.GCSType,
		ProjectID:               appConfig.GCSProjectID,
		PrivateKeyID:            appConfig.GCSPrivateKeyID,
		PrivateKey:              privateKeyPEM,
		ClientEmail:             appConfig.GCSClientEmail,
		ClientID:                appConfig.GCSClientID,
		AuthURI:                 appConfig.GCSAuthURI,
		TokenURI:                appConfig.GCSTokenURI,
		AuthProviderX509CertURL: appConfig.GCSAuthProviderX509CertURL,
		ClientX509CertURL:       appConfig.GCSClientX509CertURL,
		UniverseDomain:          appConfig.GCSUniverseDomain,
	}
	gcsClient := gcs.NewGSClient(
		gcsServiceAccount,
		appConfig.GCSBucketName,
	)
	return gcsClient, nil
}
