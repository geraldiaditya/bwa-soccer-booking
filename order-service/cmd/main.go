package cmd

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"order-service/clients"
	"order-service/common/response"
	"order-service/config"
	"order-service/constants"
	controllers "order-service/controllers/http"
	kafka2 "order-service/controllers/kafka"
	kafka "order-service/controllers/kafka/config"
	"order-service/domain/models"
	"order-service/middlewares"
	"order-service/repositories"
	"order-service/routes"
	"order-service/services"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	if err := command.Execute(); err != nil {
		panic(err)
	}
}

var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
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
			&models.Order{},
			&models.OrderHistory{},
			&models.OrderField{},
		)
		client := clients.NewClientRegistry()
		repository := repositories.NewRepositoryRegistry(db)
		service := services.NewServiceRegistry(repository, client)

		controller := controllers.NewControllerRegistry(service)
		serveHttp(controller, client)
		serveKafkaConsumer(service)
	},
}

func serveHttp(controller controllers.IControllerRegistry, client clients.IClientRegistry) {
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
			Message: "Welcome to Order Service",
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

	go func() {
		port := fmt.Sprintf(":%d", config.Config.Port)
		router.Run(port)
	}()
}

func serveKafkaConsumer(service services.IServiceRegistry) {
	kafkaConsumerConfig := sarama.NewConfig()
	kafkaConsumerConfig.Consumer.MaxWaitTime = time.Duration(config.Config.Kafka.MaxWaitTimeInMs) * time.Millisecond
	kafkaConsumerConfig.Consumer.MaxProcessingTime = time.Duration(config.Config.Kafka.MaxProcessingTimeInMs) * time.Millisecond
	kafkaConsumerConfig.Consumer.Retry.Backoff = time.Duration(config.Config.Kafka.BackoffTimeInMS) * time.Millisecond
	kafkaConsumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	kafkaConsumerConfig.Consumer.Offsets.AutoCommit.Enable = true
	kafkaConsumerConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	kafkaConsumerConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}

	brokers := config.Config.Kafka.Brokers
	groupId := config.Config.Kafka.GroupID
	topics := config.Config.Kafka.Topics

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupId, kafkaConsumerConfig)
	if err != nil {
		logrus.Errorf("Error creating consumer group: %v", err)
		return
	}
	defer consumerGroup.Close()

	consumer := kafka.NewConsumerGroup()
	kafkaRegistry := kafka2.NewKafkaRegistry(service)
	kafkaConsumer := kafka.NewKafkaConsumer(consumer, kafkaRegistry)

	kafkaConsumer.Register()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			err = consumerGroup.Consume(ctx, topics, consumer)
			if err != nil {
				logrus.Errorf("failed to consume: %v", err)
				panic(err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	logrus.Infof("kafka consumer started")
	<-signals
	logrus.Infof("kafka consumer shutting down")
}
