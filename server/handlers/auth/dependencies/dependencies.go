package dependencies

import (
	"github.com/gin-gonic/gin"
	"git.zam.io/wallet-backend/web-api/config/server"
	"git.zam.io/wallet-backend/web-api/db"
	"git.zam.io/wallet-backend/web-api/services/nosql"
	"git.zam.io/wallet-backend/web-api/services/notifications"
	"git.zam.io/wallet-backend/web-api/services/sessions"
	"go.uber.org/dig"
)

// Dependencies dependencies used by auth and signup endpoints
type Dependencies struct {
	dig.In

	Routes         gin.IRouter `name:"api_routes"`
	Db             *db.Db
	SessStorage    sessions.IStorage
	Notificator    notifications.ISender
	AuthMiddleware gin.HandlerFunc `name:"auth"`
	Generator      notifications.IGenerator
	Storage        nosql.IStorage

	Conf server.Scheme
}
