package auth

import (
	"git.zam.io/wallet-backend/web-api/internal/server/handlers/auth/dependencies"
	"git.zam.io/wallet-backend/web-api/internal/server/handlers/auth/recovery"
	"git.zam.io/wallet-backend/web-api/internal/server/handlers/auth/signup"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	"github.com/gin-gonic/gin"
)

// Register creates and registers /auth routes with given dependencies
func Register(deps dependencies.Dependencies) gin.IRouter {
	// placed here until more user endpoints come
	deps.Routes.GET("/user/me", deps.AuthMiddleware, base.WrapHandler(StatFactory(deps.Db, deps.StatsGetter)))

	group := deps.Routes.Group("/auth")

	group.POST("/signin", base.WrapHandler(SigninHandlerFactory(
		deps.Db, deps.SessStorage, deps.Conf.Auth.TokenExpire,
	)))

	group.DELETE("/signout", deps.AuthMiddleware, base.WrapHandler(SignoutHandlerFactory(
		deps.SessStorage, deps.Conf.Auth.TokenName,
	)))

	group.GET("/refresh_token", deps.AuthMiddleware, base.WrapHandler(RefreshTokenHandlerFactory(
		deps.SessStorage, deps.Conf.Auth.TokenName, deps.Conf.Auth.TokenExpire,
	)))

	group.GET("/check", deps.AuthMiddleware, base.WrapHandler(CheckHandlerFactory()))

	// register signup endpoints
	signup.Register(group.Group("/signup"), deps)

	// register recovery endpoints
	recovery.Register(group.Group("/recovery"), deps)

	return group
}
