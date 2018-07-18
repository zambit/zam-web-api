package auth

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/auth/dependencies"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/auth/signup"
)

// Register creates and registers /auth routes with given dependencies
func Register(deps dependencies.Dependencies) gin.IRouter {
	group := deps.Routes.Group("/auth")

	group.POST("/signin", base.WrapHandler(SigninHandlerFactory(
		deps.Db, deps.SessStorage, deps.Conf.Auth.TokenExpire,
	)))

	group.DELETE("/signout", deps.AuthMiddleware, base.WrapHandler(SignoutHandlerFactory(
		deps.SessStorage, deps.Conf.Auth.TokenName,
	)))

	group.POST("/refresh_token", deps.AuthMiddleware, base.WrapHandler(RefreshTokenHandlerFactory(
		deps.SessStorage, deps.Conf.Auth.TokenName,
	)))

	group.GET("/check", deps.AuthMiddleware, base.WrapHandler(CheckHandlerFactory()))

	// register signup endpoints
	signup.Register(group.Group("/signup"), deps)

	return group
}
