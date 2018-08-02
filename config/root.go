package config

import (
	"git.zam.io/wallet-backend/web-api/config/db"
	"git.zam.io/wallet-backend/web-api/config/server"
	"github.com/spf13/viper"
	"time"
	"git.zam.io/wallet-backend/web-api/config/isc"
)

// RootScheme is the scheme used by top-level app
type RootScheme struct {
	// Env describes current environment
	Env string

	// DB connection description
	DB db.Scheme

	// Server holds different web-server related configuration values
	Server server.Scheme

	// ISC contains inter-process communication params
	ISC isc.Scheme
}

// Init set default values
func Init(v *viper.Viper) {
	v.SetDefault("Env", "test")
	v.SetDefault("Db.Uri", "postgresql://postgres:postgres@localhost:5432/postgres")
	v.SetDefault("Server.Host", "localhost")
	v.SetDefault("Server.Port", 9999)
	v.SetDefault("Server.Auth.TokenExpire", time.Hour*24)
	v.SetDefault("Server.Auth.TokenName", "Bearer")
	v.SetDefault("Server.Auth.SignUpTokenExpire", time.Hour*24)
	v.SetDefault("Server.Auth.SignUpRetryDelay", time.Minute)
	v.SetDefault("Server.Storage.URI", "mem://")
	v.SetDefault("Server.Generator.CodeLen", 6)
	v.SetDefault("Server.Generator.CodeAlphabet", "1234567890")
}
