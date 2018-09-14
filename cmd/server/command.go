package server

import (
	"fmt"
	"git.zam.io/wallet-backend/common/pkg/types"
	"git.zam.io/wallet-backend/web-api/cmd/utils"
	"git.zam.io/wallet-backend/web-api/config"
	dbconf "git.zam.io/wallet-backend/web-api/config/db"
	iscconf "git.zam.io/wallet-backend/web-api/config/isc"
	"git.zam.io/wallet-backend/web-api/config/logging"
	serverconf "git.zam.io/wallet-backend/web-api/config/server"
	internalproviders "git.zam.io/wallet-backend/web-api/internal/providers"
	_ "git.zam.io/wallet-backend/web-api/internal/server/handlers"
	"git.zam.io/wallet-backend/web-api/internal/server/handlers/auth"
	"git.zam.io/wallet-backend/web-api/internal/server/handlers/kyc"
	"git.zam.io/wallet-backend/web-api/pkg/providers"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

// Create and initialize server command for given viper instance
func Create(v *viper.Viper, cfg *config.RootScheme) cobra.Command {
	command := cobra.Command{
		Use:   "server",
		Short: "Runs Web-API server",
		RunE: func(_ *cobra.Command, args []string) error {
			return serverMain(*cfg)
		},
	}
	// add common flags
	command.Flags().StringP("server.host", "l", v.GetString("server.host"), "host to serve on")
	command.Flags().IntP("server.port", "p", v.GetInt("server.port"), "port to serve on")
	command.Flags().String(
		"db.uri",
		v.GetString("db.uri"),
		"postgres connection uri",
	)
	v.BindPFlags(command.Flags())

	return command
}

// serverMain
func serverMain(cfg config.RootScheme) (err error) {
	// create DI container and populate it with providers
	c := dig.New()

	// provide container itself
	utils.MustProvide(c, func() *dig.Container {
		return c
	})

	// provide configuration and her parts
	utils.MustProvide(c, func() (config.RootScheme, dbconf.Scheme, iscconf.Scheme, serverconf.Scheme, logging.Scheme, types.Environment) {
		return cfg, cfg.DB, cfg.ISC, cfg.Server, cfg.Logging, cfg.Env
	})

	// provide root logger
	utils.MustProvide(c, providers.RootLogger)

	// provide sentry reporter
	utils.MustProvide(c, providers.Reporter)

	// provide ordinal db connection
	utils.MustProvide(c, providers.DB)

	// provide nosql storage
	utils.MustProvide(c, providers.Storage)

	// provide sessions storage
	utils.MustProvide(c, providers.SessionsStorage)

	// provide static generator
	utils.MustProvide(c, providers.Generator)

	// provide broker
	utils.MustProvide(c, providers.Broker)

	// provide user wallets stat getter
	utils.MustProvide(c, internalproviders.UserWalletStatsGetter)

	// provide old notificator
	utils.MustProvide(c, internalproviders.Notificator)

	// provide gin engine
	utils.MustProvide(c, providers.GinEngine)
	utils.MustProvide(c, providers.RootRouter, dig.Name("root"))

	// provide events notificator
	utils.MustProvide(c, internalproviders.EventNotificator)

	// provide api router
	utils.MustProvide(c, internalproviders.ApiRoutes, dig.Name("api_routes"))

	// provide auth middleware
	utils.MustProvide(c, providers.AuthMiddleware, dig.Name("auth"))

	// register handlers
	utils.MustInvoke(c, static.Register)
	utils.MustInvoke(c, auth.Register)
	utils.MustInvoke(c, kyc.Register)

	// Run server!
	utils.MustInvoke(c, func(engine *gin.Engine) error {
		return engine.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	})

	return
}
