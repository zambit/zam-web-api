package config

import (
	"github.com/spf13/viper"
	"gitlab.com/ZamzamTech/wallet-api/config/db"
	"gitlab.com/ZamzamTech/wallet-api/config/server"
	"time"
)

// RootScheme is the scheme used by top-level app
type RootScheme struct {
	// Env describes current environment
	Env string

	// DB connection description
	DB db.Scheme

	// Server holds different web-server related configuration values
	Server server.Scheme
}

// Init set default values
func Init(v *viper.Viper) {
	v.SetDefault("Env", "test")
	v.SetDefault("Db.Uri", "postgresql://postgres:postgres@localhost:5432/postgres")
	v.SetDefault("Server.Host", "localhost")
	v.SetDefault("Server.Port", 9999)
	v.SetDefault("Server.Auth.TokenExpire", time.Hour*24)
	v.SetDefault("Server.Auth.TokenName", "Bearer")
}
