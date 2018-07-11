package auth

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
	"go.uber.org/dig"
	"gitlab.com/ZamzamTech/wallet-api/config/server"
)

// Dependencies dependencies used by
type Dependencies struct {
	dig.In

	Routes         gin.IRouter `name:"api_routes"`
	Db             *db.Db
	SessStorage    sessions.IStorage
	Notificator    notifications.ISender
	AuthMiddleware gin.HandlerFunc `name:"auth"`

	Conf           server.Scheme
}

// Register creates and registers /auth routes with given dependencies
func Register(deps Dependencies) gin.IRouter {
	group := deps.Routes.Group("/auth")
	group.POST("/signup", base.WrapHandler(SignupHandlerFactory(
		deps.Db, deps.SessStorage, deps.Notificator, deps.Conf.Auth.TokenExpire,
	)))
	return group
}