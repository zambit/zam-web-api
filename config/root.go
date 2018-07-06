package config

import (
	"gitlab.com/ZamzamTech/wallet-api/config/db"
	"github.com/spf13/viper"
	"gitlab.com/ZamzamTech/wallet-api/config/server"
)

// RootScheme is the scheme used by top-level app
type RootScheme struct {
	// Env describes current environment
	Env string

	// DB connection
	DB db.Scheme

	// Server
	Server server.Scheme
}

// Init set default values
func Init(v *viper.Viper) {
	v.SetDefault("env", "test")
	v.SetDefault("db.uri", "postgresql://postgres:postgres@localhost:5433/postgres")
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 9999)
}