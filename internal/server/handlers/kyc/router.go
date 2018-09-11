package kyc

import (
	"git.zam.io/wallet-backend/web-api/db"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// Dependencies dependencies used by auth and signup endpoints
type Dependencies struct {
	dig.In

	Db             *db.Db
	Routes         gin.IRouter     `name:"api_routes"`
	AuthMiddleware gin.HandlerFunc `name:"auth"`
}

// Register
func Register(deps Dependencies) {
	group := deps.Routes.Group("/user/me/personal", deps.AuthMiddleware)
	group.POST("", base.WrapHandler(CreateFactory(deps.Db)))
	group.GET("", base.WrapHandler(GetFactory(deps.Db)))
}
