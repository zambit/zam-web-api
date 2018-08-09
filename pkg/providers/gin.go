package providers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GinEngine
func GinEngine(logger logrus.FieldLogger) *gin.Engine {
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowMethods = append(corsCfg.AllowMethods, "DELETE", "PATCH")
	corsCfg.AllowAllOrigins = true
	corsCfg.AllowHeaders = append(
		corsCfg.AllowHeaders, "Authorization", "Accept-Encoding", "X-CSRF-Token", "Accept",
	)
	corsCfg.AllowCredentials = true

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
