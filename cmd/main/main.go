package main

import (
	"github.com/spf13/viper"
	"git.zam.io/wallet-backend/web-api/cmd/root"
	"git.zam.io/wallet-backend/web-api/cmd/server"
	"git.zam.io/wallet-backend/web-api/config"
)

// main executes specified command using cobra, on error will panic for nice stack print and non-zero exit code
func main() {
	var cfg config.RootScheme
	v := viper.New()

	config.Init(v)
	rootCmd := root.Create(v, &cfg)
	serverCmd := server.Create(v, &cfg)
	serverCmd.Flags().AddFlagSet(rootCmd.Flags())
	rootCmd.AddCommand(&serverCmd)

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
