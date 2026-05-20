package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/IBM/sarama"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"order-service/clients"
	"order-service/common/response"
	"order-service/config"
	"order-service/constants"
	healthController "order-service/controllers/health"
	controllers "order-service/controllers/http"
	kafka2 "order-service/controllers/kafka"
	kafka "order-service/controllers/kafka/config"
	_ "order-service/docs"
	"order-service/domain/models"
	"order-service/middlewares"
	"order-service/repositories"
	healthRoute "order-service/routes/health"
	"order-service/routes"
	"order-service/services"
)

// @title Order Service API
// @version 1.0
// @description Microservice for managing soccer booking orders and payments.
// @host localhost:8003
// @BasePath /api/v1

func Run() {
	if err := command.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()
		if err := config.Config.Validate(); err != nil {
			logrus.Fatalf("invalid configuration: %v", err)
		}

		db, err := config.InitDatabase()
		if err != nil {
			logrus.Fatal(err)
		}
		err = db.AutoMigrate(
			&models.Order{},
			&models.OrderHistory{},
			&models.OrderField{},
		)
		if err != nil {
			logrus.Fatal(err)
		}

		client := clients.NewClientRegistry()
		repository := repositories.NewRepositoryRegistry(db)
		service := services.NewServiceRegistry(repository, client)
		controller := controllers.NewControllerRegistry(service)

		router := gin.Default()
		// HandlePanic must be first so it recovers panics from all subsequent middleware/handlers.
		router.Use(middlewares.HandlePanic())
		router.Use(middlewares.RequestLogger())
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
				Message: "Welcome to Order Service",
			})
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

		serverErr := make(chan error, 1)
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				serverErr <- err
			}
		}()
		logrus.Infof("HTTP server started on port %s", port)

		kafkaConsumerConfig := sarama.NewConfig()
		kafkaConsumerConfig.Consumer.MaxWaitTime = time.Duration(config.Config.Kafka.MaxWaitTimeInMs) * time.Millisecond
		kafkaConsumerConfig.Consumer.MaxProcessingTime = time.Duration(config.Config.Kafka.MaxProcessingTimeInMs) * time.Millisecond
		kafkaConsumerConfig.Consumer.Retry.Backoff = time.Duration(config.Config.Kafka.BackoffTimeInMS) * time.Millisecond
		kafkaConsumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
		kafkaConsumerConfig.Consumer.Offsets.AutoCommit.Enable = false
		kafkaConsumerConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
			sarama.NewBalanceStrategyRoundRobin(),
		}

		brokers := config.Config.Kafka.Brokers
		groupId := config.Config.Kafka.GroupID
		topics := config.Config.Kafka.Topics

		consumerGroup, err := sarama.NewConsumerGroup(brokers, groupId, kafkaConsumerConfig)
		if err != nil {
			logrus.Fatalf("error creating consumer group: %v", err)
		}

		consumer := kafka.NewConsumerGroup()
		kafkaRegistry := kafka2.NewKafkaRegistry(service)
		kafkaConsumer := kafka.NewKafkaConsumer(consumer, kafkaRegistry)
		kafkaConsumer.Register()

		shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
		defer shutdownCancel()

		backoff := time.Duration(config.Config.Kafka.BackoffTimeInMS) * time.Millisecond

		go func() {
			for {
				if consumeErr := consumerGroup.Consume(shutdownCtx, topics, consumer); consumeErr != nil {
					logrus.Errorf("failed to consume: %v", consumeErr)
					// backoff before retrying to avoid spinning on persistent broker errors
					select {
					case <-time.After(backoff):
					case <-shutdownCtx.Done():
						return
					}
				}
				if shutdownCtx.Err() != nil {
					return
				}
			}
		}()
		logrus.Infof("kafka consumer started")

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-signals:
			logrus.Infof("received signal %v, initiating shutdown", sig)
		case err := <-serverErr:
			logrus.Errorf("HTTP server error: %v, initiating shutdown", err)
		}

		shutdownCancel()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logrus.Errorf("server shutdown error: %v", err)
		}

		if err := consumerGroup.Close(); err != nil {
			logrus.Errorf("consumer group close error: %v", err)
		}

		if sqlDB, err := db.DB(); err != nil {
			logrus.Errorf("failed to get underlying sql.DB: %v", err)
		} else {
			sqlDB.Close()
		}

		logrus.Info("server exiting")
	},
}
