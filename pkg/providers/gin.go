package providers

import (
	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

// GinEngine
func GinEngine(logger logrus.FieldLogger) *gin.Engine {
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
}

// RootRouter
func RootRouter(engine *gin.Engine) gin.IRouter {
	return engine
}