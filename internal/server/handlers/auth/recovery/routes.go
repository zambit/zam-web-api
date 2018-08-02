package recovery

import (
	"git.zam.io/wallet-backend/web-api/internal/server/handlers/auth/dependencies"
	"git.zam.io/wallet-backend/web-api/internal/server/handlers/base"
	"github.com/gin-gonic/gin"
)

// Register creates and registers /auth/recovery routes with given dependencies
func Register(group gin.IRouter, deps dependencies.Dependencies) gin.IRouter {
	group.POST("/start", base.WrapHandler(StartHandlerFactory(
		deps.Db, deps.Notificator, deps.Generator, deps.Storage,
		deps.Conf.Auth.SignUpTokenExpire,
	)))

	group.POST("/verify", base.WrapHandler(VerifyHandlerFactory(
		deps.Db, deps.Generator, deps.Storage,
		deps.Conf.Auth.SignUpTokenExpire,
	)))

	group.PUT("/finish", base.WrapHandler(FinishHandlerFactory(deps.Db, deps.Storage, deps.Notificator)))

	return group
}
