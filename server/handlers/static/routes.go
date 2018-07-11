package static

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
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
