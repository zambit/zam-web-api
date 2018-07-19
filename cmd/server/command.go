package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/ZamzamTech/wallet-api/config"
	serverconf "gitlab.com/ZamzamTech/wallet-api/config/server"
	"gitlab.com/ZamzamTech/wallet-api/db"
	_ "gitlab.com/ZamzamTech/wallet-api/server/handlers"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/auth"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/static"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications/stub"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions/mem"
	memnosql "gitlab.com/ZamzamTech/wallet-api/services/nosql/mem"
	"go.uber.org/dig"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	"gitlab.com/ZamzamTech/wallet-api/server/middlewares"
	"github.com/gin-contrib/cors"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/auth/dependencies"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications"
)

// Create and initialize server command for given viper instance
func Create(v *viper.Viper, cfg *config.RootScheme) cobra.Command {
	command := cobra.Command{
		Use:   "server",
		Short: "Runs Wallet-API server",
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

	// provide root logger
	err = c.Provide(func() logrus.FieldLogger {
		return logrus.New()
	})
	if err != nil {
		return
	}

	// provide ordinal db connection
	err = c.Provide(db.Factory(cfg.DB.URI))
	if err != nil {
		return
	}

	// provide sessions storage
	err = c.Provide(mem.New)
	if err != nil {
		return
	}

	// provide nosql storage
	err = c.Provide(memnosql.New)
	if err != nil {
		return
	}

	// provide static generator
	err = c.Provide(func(conf serverconf.Scheme) notifications.IGenerator {
		return notifications.NewWithCodeAlphabet(conf.Generator.CodeLen, conf.Generator.CodeAlphabet)
	})
	if err != nil {
		return
	}

	// provide notificator
	err = c.Provide(stub.New)
	if err != nil {
		return
	}

	err = c.Provide(func() serverconf.Scheme {
		return cfg.Server
	})
	if err != nil {
		return
	}

	// provide gin engine
	err = c.Provide(func(logger logrus.FieldLogger) *gin.Engine {
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowMethods = append(corsCfg.AllowMethods, "DELETE")
		corsCfg.AllowAllOrigins = true
		corsCfg.AllowHeaders = []string{"*"}

		engine := gin.New()
		engine.Use(
			gin.Recovery(),
			gin.Logger(),
			cors.New(corsCfg),
		)
		return engine
	})
	if err != nil {
		return
	}

	// provide api router
	err = c.Provide(func(engine *gin.Engine) gin.IRouter {
		return engine.Group("/api/v1")
	}, dig.Name("api_routes"))
	if err != nil {
		return
	}

	// provide auth middleware
	err = c.Provide(func(sessStorage sessions.IStorage) gin.HandlerFunc {
		return middlewares.AuthMiddlewareFactory(sessStorage, cfg.Server.Auth.TokenName)
	}, dig.Name("auth"))

	// Run server!
	err = c.Invoke(func(engine *gin.Engine, dependencies dependencies.Dependencies) error {
		auth.Register(dependencies)
		static.Register(engine)
		return engine.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	})
	return
}
