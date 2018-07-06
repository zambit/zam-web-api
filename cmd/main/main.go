package main

import (
	"gitlab.com/ZamzamTech/wallet-api/cmd/root"
	"github.com/spf13/viper"
	"gitlab.com/ZamzamTech/wallet-api/cmd/server"
	"gitlab.com/ZamzamTech/wallet-api/config"
)

// main executes specified command using cobra, on error will panic for nice stack print and non-zero exit code
func main() {
	v := viper.New()

	config.Init(v)
	rootCmd := root.Create(v)
	serverCmd := server.Create(v)
	rootCmd.AddCommand(&serverCmd)

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
