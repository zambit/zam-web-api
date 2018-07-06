package server

import (
	"github.com/spf13/viper"
	"github.com/spf13/cobra"
)

// Create and initialize server command for given viper instance
func Create(v *viper.Viper) cobra.Command {
	command := cobra.Command{
		Use: "server",
		Short: "Runs Wallet-API server",
	}
	// add common flags
	command.Flags().StringP("server.host", "l", "localhost", "host to serve on")
	command.Flags().StringP("server.port", "p", "port", "port to serve on")
	command.Flags().String(
		"db.uri",
		"postgresql://postgres:postgres@localhost:5433/postgres",
		"postgres connection uri",
	)
	v.BindPFlags(command.Flags())

	return command
}