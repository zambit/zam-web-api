package providers

import (
	"git.zam.io/wallet-backend/web-api/config/server"
	"git.zam.io/wallet-backend/web-api/pkg/server/middlewares"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions"
	"github.com/gin-gonic/gin"
)

// Auth middleware
func AuthMiddleware(sessStorage sessions.IStorage, conf server.Scheme) gin.HandlerFunc {
	return middlewares.AuthMiddlewareFactory(sessStorage, conf.Auth.TokenName)
}
