package dependencies

import (
	"git.zam.io/wallet-backend/web-api/config/server"
	"git.zam.io/wallet-backend/web-api/db"
	"git.zam.io/wallet-backend/web-api/internal/services/isc"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"git.zam.io/wallet-backend/web-api/internal/services/stats"
	"git.zam.io/wallet-backend/web-api/pkg/services/nosql"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// Dependencies dependencies used by auth and signup endpoints
type Dependencies struct {
	dig.In

	Routes         gin.IRouter `name:"api_routes"`
	Db             *db.Db
	SessStorage    sessions.IStorage
	Notificator    isc.IEventNotificator
	AuthMiddleware gin.HandlerFunc `name:"auth"`
	Generator      notifications.IGenerator
	Storage        nosql.IStorage
	StatsGetter    stats.IUserWalletsGetter

	Conf server.Scheme
}
