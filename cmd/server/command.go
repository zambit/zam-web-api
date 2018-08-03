package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"git.zam.io/wallet-backend/web-api/config"
	serverconf "git.zam.io/wallet-backend/web-api/config/server"
	iscconf "git.zam.io/wallet-backend/web-api/config/isc"
	dbconf "git.zam.io/wallet-backend/web-api/config/db"
	_ "git.zam.io/wallet-backend/web-api/internal/server/handlers"
	"git.zam.io/wallet-backend/web-api/internal/server/handlers/auth"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/static"
	"git.zam.io/wallet-backend/web-api/cmd/utils"
	"git.zam.io/wallet-backend/web-api/pkg/providers"
	internalproviders "git.zam.io/wallet-backend/web-api/internal/providers"
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
	utils.MustProvide(c, func() (config.RootScheme, dbconf.Scheme, iscconf.Scheme, serverconf.Scheme) {
		return cfg, cfg.DB, cfg.ISC, cfg.Server
	})

	// provide root logger
	utils.MustProvide(c, providers.RootLogger)

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

	// Run server!
	utils.MustInvoke(c, func(engine *gin.Engine) error {
		return engine.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	})

	return
}
