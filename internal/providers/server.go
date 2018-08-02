package providers

import (
	"github.com/gin-gonic/gin"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions"
	"git.zam.io/wallet-backend/web-api/internal/server/middlewares"
	"git.zam.io/wallet-backend/web-api/config/server"
)

// Auth middleware
func AuthMiddleware(sessStorage sessions.IStorage, conf server.Scheme) gin.HandlerFunc {
	return middlewares.AuthMiddlewareFactory(sessStorage, conf.Auth.TokenName)
}
