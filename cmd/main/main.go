package main

import (
	"git.zam.io/wallet-backend/common/pkg/gitversion"
	"git.zam.io/wallet-backend/web-api/cmd/root"
	"git.zam.io/wallet-backend/web-api/cmd/server"
	"git.zam.io/wallet-backend/web-api/config"
	"github.com/spf13/viper"
)

// main executes specified command using cobra, on error will panic for nice stack print and non-zero exit code
func main() {
	// initialize version
	i := gitversion.GetInfo("0.0.1", "alpha")

	// create configuration
	cfg := config.RootScheme{
		Version: i.Version,
	}
	v := viper.New()

	config.Init(v)
	rootCmd := root.Create(v, &cfg)
	rootCmd.Long = i.BuildDescription
	serverCmd := server.Create(v, &cfg)
	rootCmd.AddCommand(&serverCmd)

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
