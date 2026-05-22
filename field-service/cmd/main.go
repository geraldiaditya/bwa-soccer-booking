package cmd

import (
	"field-service/internal/app"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// @title Field Service API
// @version 1.0
// @description Microservice for managing soccer fields and schedules.
// @host localhost:8002
// @BasePath /api/v1

var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		application, err := app.New()
		if err != nil {
			return err
		}
		return application.Run()
	},
}

func Run() {
	if err := command.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
