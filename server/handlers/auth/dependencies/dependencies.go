package dependencies

import (
	"go.uber.org/dig"
	"github.com/gin-gonic/gin"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications"
	"gitlab.com/ZamzamTech/wallet-api/services/nosql"
	"gitlab.com/ZamzamTech/wallet-api/config/server"
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
