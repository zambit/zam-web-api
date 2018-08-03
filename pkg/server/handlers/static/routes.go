package static

import (
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	"github.com/gin-gonic/gin"
)

// Register adds static routes such as not found
func Register(engine *gin.Engine) {
	engine.NoRoute(base.WrapHandler(func(c *gin.Context) (resp interface{}, code int, err error) {
		err = base.ErrorView{
			Code:    404,
			Message: "Not found",
		}
		return
	}))
}
