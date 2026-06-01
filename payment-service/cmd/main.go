package cmd

import (
	"context"
	"payment-service/internal/app"

	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Run(cmd.Context())
	},
}

func Run() {
	if err := command.ExecuteContext(context.Background()); err != nil {
		panic(err)
	}
}
