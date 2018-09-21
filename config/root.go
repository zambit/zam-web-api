package config

import (
	"git.zam.io/wallet-backend/common/pkg/types"
	"git.zam.io/wallet-backend/web-api/config/db"
	"git.zam.io/wallet-backend/web-api/config/isc"
	"git.zam.io/wallet-backend/web-api/config/logging"
	"git.zam.io/wallet-backend/web-api/config/server"
	"github.com/blang/semver"
	"github.com/spf13/viper"
	"time"
)

// RootScheme is the scheme used by top-level app
type RootScheme struct {
	// Version current version, default is set during build process
	Version semver.Version

	// Env describes current environment
	Env types.Environment

	// DB connection description
	DB db.Scheme

	// Server holds different web-server related configuration values
	Server server.Scheme

	// ISC contains inter-process communication params
	ISC isc.Scheme

	// Logging logging configuration
	Logging logging.Scheme
}

// Init set default values
func Init(v *viper.Viper) {
	v.SetDefault("Env", "test")
	v.SetDefault("Logging.LogLevel", "info")
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
