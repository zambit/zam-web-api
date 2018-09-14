package providers

import (
	"git.zam.io/wallet-backend/common/pkg/types"
	"git.zam.io/wallet-backend/web-api/pkg/services/sentry"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GinEngine
func GinEngine(env types.Environment, logger logrus.FieldLogger, reporter sentry.IReporter) *gin.Engine {
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowMethods = append(corsCfg.AllowMethods, "DELETE", "PATCH")
	corsCfg.AllowAllOrigins = true
	corsCfg.AllowHeaders = append(
		corsCfg.AllowHeaders, "Authorization", "Accept-Encoding", "X-CSRF-Token", "Accept",
	)
	corsCfg.AllowCredentials = true

	gin.SetMode(coerceEnvToGin(env))
	// set global reporter instance
	sentry.SetGlobal(reporter)

	logger.Warnf(
		"ATTENTION: gin will print below stupid message about current environment, don't trust, real values is: %s",
		env,
	)
	engine := gin.New()
	engine.Use(
		gin.Recovery(),
		gin.Logger(),
		cors.New(corsCfg),
	)
	return engine
}

// RootRouter
func RootRouter(engine *gin.Engine) gin.IRouter {
	return engine
}

func coerceEnvToGin(env types.Environment) string {
	if env.IsProduction() {
		return "release"
	} else {
		return "debug"
	}
}
