// @title           User Service API
// @version         1.0
// @description     User management and authentication service.
// @host            localhost:8001
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "user-service/docs"

	"user-service/common/response"
	"user-service/config"
	"user-service/constants"
	"user-service/controllers"
	"user-service/database/seeders"
	"user-service/domain/models"
	"user-service/middlewares"
	"user-service/repositories"
	"user-service/routes"
	"user-service/services"
)

var rootCmd = &cobra.Command{
	Use:   "user-service",
	Short: "User Service CLI",
}

var serveCmd = &cobra.Command{
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

		repository := repositories.NewRepositoryRegistry(db)
		service := services.NewServiceRegistry(repository)

		controller := controllers.NewControllerRegistry(service)

		router := gin.Default()
		router.Use(middlewares.HandlePanic())
		router.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-apikey, x-request-at")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
			c.Next()
		})
		router.OPTIONS("/*any", func(c *gin.Context) {
			c.AbortWithStatus(http.StatusNoContent)
		})
		router.Use(middlewares.ErrorHandler())

		router.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, response.Response{
				Status:  constants.Error,
				Message: fmt.Sprintf("Path %s", http.StatusText(http.StatusNotFound)),
			})
		})
		router.GET("/", func(context *gin.Context) {
			context.JSON(http.StatusOK, response.Response{
				Status:  constants.Success,
				Message: "Welcome to User Service",
			})
		})
		router.GET("/health", func(context *gin.Context) {
			context.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		lmt := tollbooth.NewLimiter(
			config.Config.RateLimiterMaxRequest,
			&limiter.ExpirableOptions{
				DefaultExpirationTTL: time.Duration(config.Config.RateLimiterTimeSecond) * time.Second,
			},
		)
		router.Use(middlewares.RateLimiter(lmt))

		group := router.Group("/api/v1")
		route := routes.NewRouteRegistry(controller, group)
		route.Serve()

		port := fmt.Sprintf(":%d", config.Config.Port)
		srv := &http.Server{Addr: port, Handler: router}

		go func() {
			if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err = srv.Shutdown(ctx); err != nil {
			panic(err)
		}
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()
		config.Init()
		db, err := config.InitDatabase()
		if err != nil {
			panic(err)
		}

		fmt.Println("Running AutoMigrate...")
		err = db.AutoMigrate(
			&models.Role{}, &models.User{},
		)
		if err != nil {
			panic(err)
		}
		fmt.Println("AutoMigrate completed.")
	},
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Run database seeders",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()
		config.Init()
		db, err := config.InitDatabase()
		if err != nil {
			panic(err)
		}

		fmt.Println("Running Seeders...")
		seeders.NewSeederRegistry(db).Run()
		fmt.Println("Seeders completed.")
	},
}

func Run() {
	rootCmd.AddCommand(serveCmd, migrateCmd, seedCmd)
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
