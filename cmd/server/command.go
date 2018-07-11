package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/ZamzamTech/wallet-api/config"
	serverconf "gitlab.com/ZamzamTech/wallet-api/config/server"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/auth"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/static"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications/stub"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions/mem"
	"go.uber.org/dig"
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
	command.Flags().StringP("server.host", "l", "localhost", "host to serve on")
	command.Flags().StringP("server.port", "p", "port", "port to serve on")
	command.Flags().String(
		"db.uri",
		"postgresql://postgres:postgres@localhost:5433/postgres",
		"postgres connection uri",
	)
	v.BindPFlags(command.Flags())

	return command
}

// ginValidatorV9 overrides default gin validator
type ginValidatorV9 struct {
	validator *validator.Validate
}

func (v ginValidatorV9) ValidateStruct(val interface{}) error {
	return v.validator.Struct(val)
}

func (v ginValidatorV9) Engine() interface{} {
	return v.validator
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

	err = c.Provide(func() gin.HandlerFunc {
		return nil
	}, dig.Name("auth"))
	if err != nil {
		return
	}

	// provide gin engine
	err = c.Provide(func(logger logrus.FieldLogger) *gin.Engine {
		engine := gin.New()
		engine.Use(
			gin.Recovery(),
			gin.Logger(),
		)
		binding.Validator = ginValidatorV9{validator: validator.New()}
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

	// Run server!
	err = c.Invoke(func(engine *gin.Engine, dependencies auth.Dependencies) error {
		auth.Register(dependencies)
		static.Register(engine)
		return engine.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	})
	return
}
